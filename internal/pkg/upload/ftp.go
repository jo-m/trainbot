package upload

import (
	"context"
	"fmt"
	"io"
	"net/textproto"
	"path"
	"strings"

	"github.com/jlaffaye/ftp"
)

// FTPConfig is the configuration to connect to a FTP server.
type FTPConfig struct {
	Host     string `arg:"--upload-ftp-host,env:UPLOAD_FTP_HOST" help:"FTP hostname" placeholder:"HOST"`
	Port     uint16 `arg:"--upload-ftp-port,env:UPLOAD_FTP_PORT" help:"FTP port" default:"21" placeholder:"PORT"`
	User     string `arg:"--upload-ftp-user,env:UPLOAD_FTP_USER" help:"FTP username" placeholder:"USER"`
	Password string `arg:"--upload-ftp-password,env:UPLOAD_FTP_PASSWORD" help:"FTP password" placeholder:"PASS"`
	PWD      string `arg:"--upload-ftp-pwd,env:UPLOAD_FTP_PWD" help:"FTP working directory to change to, expected to exist" default:"." placeholder:"DIR"`
}

// FTP is a FTP uploader. Use NewFTP to create an instance.
type FTP struct {
	conf FTPConfig
	conn *ftp.ServerConn
}

// Compile time interface check.
var _ Uploader = (*FTP)(nil)

// NewFTP connects and authenticates to an FTP server.
func NewFTP(ctx context.Context, c FTPConfig) (*FTP, error) {
	conn, err := ftp.Dial(fmt.Sprintf("%s:%d", c.Host, c.Port), ftp.DialWithContext(ctx))
	if err != nil {
		return nil, err
	}

	err = conn.Login(c.User, c.Password)
	if err != nil {
		_ = conn.Quit()
		return nil, err
	}

	err = conn.ChangeDir(c.PWD)
	if err != nil {
		_ = conn.Quit()
		return nil, err
	}

	return &FTP{
		conf: c,
		conn: conn,
	}, nil
}

// Close implements Uploader.
func (f *FTP) Close() error {
	return f.conn.Quit()
}

func isFTPErr(err error, code int) bool {
	if errF, ok := err.(*textproto.Error); ok {
		return errF.Code == code
	}
	return false
}

func (f *FTP) createDirs(dirsPath string) error {
	components := strings.Split(dirsPath, "/")

	for i := 0; i < len(components); i++ {
		dir := path.Join(components[:i+1]...)
		err := f.conn.MakeDir(dir)
		if err != nil {
			if isFTPErr(err, 550) {
				continue
			}
			return err
		}
	}

	return nil
}

// Upload implements Uploader.
func (f *FTP) Upload(_ context.Context, remotePath string, contents io.Reader) error {
	err := f.createDirs(path.Dir(remotePath))
	if err != nil {
		return err
	}

	return f.conn.Stor(remotePath, contents)
}

// AtomicUpload implements Uploader.
func (f *FTP) AtomicUpload(ctx context.Context, remotePath string, contents io.Reader) error {
	tempName := remotePath + ".__temp__"
	err := f.Upload(ctx, tempName, contents)
	if err != nil {
		return err
	}

	return f.conn.Rename(tempName, remotePath)
}
