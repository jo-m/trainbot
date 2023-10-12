# Onlytrains

<img src="frontend/src/assets/logo-day.svg" height="100" width="100">

Watches a piece of train track, detects passing trains, and stitches together images of them.
Should work with any video4linux USB cam, or Raspberry Pi camera v3 modules.

Frontend: <https://trains.jo-m.ch/>

[A collection of some "special" sightings.](https://trains.jo-m.ch/#/trains/list?filter=%257B%2522where%2522%253A%257B%2522favs%2522%253A%2522id%2520IN%2520(577%252C2405%252C2320%252C2193%252C1342%252C1039%252C343%252C407%252C350%252C307%252C1724%252C887%252C2485%252C3002%252C2950%252C2949%252C2896%252C2870%252C2853%252C2839%252C2827%252C2815%252C2802%252C3403%252C3224%252C3008%252C2766%252C2483%252C3410%252C3425%252C3424%252C3592%252C3576%252C3986%252C3715%252C2462%252C3846%252C3903%252C3981%252C3999%252C3971%252C4045%252C4160%252C4051%252C4362%252C4300%252C4504%252C4484%252C4456%252C4669%252C4794%252C4792%252C4790%252C4796%252C4797%252C4801%252C4813%252C4814%252C4815%252C4816%252C4818%252C4820%252C4827%252C4829%252C4831%252C4841%252C4840%252C4839%252C4876%252C4874%252C4873%252C4855%252C4844%252C4894%252C4890%252C4889%252C4883%252C4882%252C5058%252C5045%252C5272%252C5257%252C5241%252C5148%252C5146%252C4823%252C5437%252C2754%252C3770%252C3768%252C4025%252C4158%252C4426%252C4430%252C5325%252C5401%252C6124%252C6567%252C6560%252C6553%252C5972%252C5535%252C5700%252C6786%252C7232%252C8332%252C8334%252C8137%252C7911%252C8532%252C8518%252C8496%252C8415%252C7956%252C7939%252C7136%252C7000%252C7001%252C9328%252C9321%252C9286%252C9281%252C9213%252C9207%252C9188%252C9649%252C9648%252C9621%252C9614%252C9584%252C9529%252C9528%252C9450%252C9422%252C9268%252C9231%252C9179%252C9175%252C8786%252C8588%252C11529%252C11420%252C11406%252C11224%252C11292%252C11223%252C11199%252C11101%252C11094%252C11027%252C11014%252C10773%252C10626%252C10349%252C10333%252C9846%252C9866%252C9814%252C12424%252C12345%252C12324%252C12216%252C12219%252C12221%252C12226%252C12235%252C12556%252C13024%252C12607%252C13267%252C13989%252C13988%252C13979%252C13914%252C13909%252C13896%252C13886%252C13728%252C13513%252C13507%252C13461%252C12410%252C12331%252C14193%252C14184%252C14213%252C14252%252C14336%252C14362%252C14373%252C13420%252C3643%252C13489%252C13460%252C13499%252C15347%252C15276%252C15263%252C15201%252C15068%252C15033%252C14985%252C14809%252C14821%252C14702%252C15443%252C15435%252C15414%252C15374%252C17103%252C17089%252C17088%252C17087%252C17084%252C17063%252C17058%252C17054%252C16954%252C16952%252C16896%252C16895%252C16890%252C16878%252C16876%252C16864%252C16856%252C16838%252C16690%252C16644%252C16388%252C16386%252C16296%252C16283%252C16282%252C16269%252C16255%252C16239%252C16152%252C16127%252C16113%252C16100%252C16086%252C16072%252C16053%252C15923%252C15885%252C15879%252C15877%252C15874%252C15783%252C15691%252C15673%252C15615%252C15577%252C15564%252C18079%252C14479%252C138%252C11810%252C18090%252C17983%252C17974%252C17962%252C17956%252C17860%252C17581%252C17536%252C17472%252C17471%252C17468%252C17305%252C17252%252C17211%252C17190%252C18288%252C18599%252C18538)%2522%257D%257D)

The name Onlytrains is credited to [@timethy](https://github.com/timethy).

[<img src="internal/pkg/stitch/testdata/set0/day.jpg">](internal/pkg/stitch/testdata/set0/day.jpg)
[<img src="internal/pkg/stitch/testdata/set0/night.jpg">](internal/pkg/stitch/testdata/set0/night.jpg)
[<img src="internal/pkg/stitch/testdata/set0/rain.jpg">](internal/pkg/stitch/testdata/set0/rain.jpg)
[<img src="internal/pkg/stitch/testdata/set0/snow.jpg">](internal/pkg/stitch/testdata/set0/snow.jpg)
[<img src="demo.gif">](demo.gif)

It also contains some packages which might be useful for other purposes:

* [pkg/pmatch](pkg/pmatch): Image patch matching
* [pkg/ransac](pkg/ransac): RANSAC algorithm implementation

The binaries are currently built and tested on X86_64 and a Raspberry Pi 4 B.

## Assumptions and notes on computer vision

The computer vision used in trainbot is fairly naive and simple.
There is no camera calibration, image stabilization, undistortion, perspective mapping, or "real" object tracking.
This allows us to stay away from complex dependencies like OpenCV, and keeps the computational requirements low.
All processing happens on CPU.

The assumptions are (there might be more implicit ones):

1. Trains only appear in a (manually) pre-cropped region.
1. The camera is stable and the image does not move around in any direction.
1. There are no large fast brightness changes.
1. Trains have a given min and max speed.
1. We are looking at the tracks more or less perpendicularly in the chosen image crop region.
1. Trains are coming from one direction at a time, crossings are not handled properly.
1. Trains have a constant acceleration (might be 0) and do not stop and turn around while in front of the camera.

## Build system

There is a helper `Makefile` which calls the standard Go build tools and an arm64 cross build inside Docker.

## V4L Settings

```bash
# list
ffmpeg -f v4l2 -list_formats all -i /dev/video2
v4l2-ctl --all --device /dev/video2

# exposure
v4l2-ctl -c exposure_auto=3 --device /dev/video2

# autofocus
v4l2-ctl -c focus_auto=1 --device /dev/video2

# fixed
v4l2-ctl -c focus_auto=0 --device /dev/video2
v4l2-ctl -c focus_absolute=0 --device /dev/video2
v4l2-ctl -c focus_absolute=1023 --device /dev/video2

ffplay -f video4linux2 -framerate 30 -video_size 3264x2448 -pixel_format mjpeg /dev/video2
ffplay -f video4linux2 -framerate 30 -video_size 1920x1080 -pixel_format mjpeg /dev/video2

ffmpeg -f v4l2 -framerate 30 -video_size 3264x2448 -pixel_format mjpeg -i /dev/video2 output.avi
```

## RasPi Cam v3 utils

```bash
# setup
sudo apt-get install libcamera0 libcamera-apps-lite
sudo apt install -y vlc

# grab frame
# https://www.raspberrypi.com/documentation/computers/camera_software.html#libcamera-and-libcamera-apps
libcamera-jpeg -o out.jpg -t 1 --width 4608 --height 2592 --rotation 180 --autofocus-mode=manual --lens-position=2
libcamera-jpeg -o out.jpg -t 1 --width 2304 --height 1296 --rotation 180 --autofocus-mode=manual --lens-position=4.5 --roi 0.25,0.5,0.5,0.5

# record video
DATE=$(date +'%F_%H-%M-%S'); libcamera-vid -o $DATE.h264 --save-pts $DATE.txt --width 1080 --height 720 --rotation 180 --autofocus-mode=manual --lens-position=0 -t 0

# stream through network
libcamera-vid -t 0 --inline --nopreview --width 4608 --height 2592 --rotation 180 --codec mjpeg --framerate 5 --listen -o tcp://0.0.0.0:8080 --autofocus-mode=manual --lens-position=0 --roi 0.25,0.5,0.5,0.5
# on localhost
ffplay http://pi4:8080/video.mjpeg

# manually record video for test cases
libcamera-vid \
   --verbose=1 \
   --timeout=0 \
   --inline \
   --nopreview \
   --width 240 --height 280 \
   --roi 0.429688,0.185185,0.104167,0.216049 \
   --mode=2304:1296:12:P \
   --framerate 30 \
   --autofocus-mode=manual --lens-position=0.000000 \
   --rotation=0 \
   -o vid.h264 --save-pts vid-timestamps.txt

mkvmerge -o test.mkv --timecodes 0:vid-timestamps.txt vid.h264
```

## Deployment

### Raspberry Pi

```bash
sudo usermod -a -G video pi

# confighelper
./confighelper-arm64 --log-pretty --input=picam3 --listen-addr=0.0.0.0:8080
```

The current production deployment is in a Tmux session...

```bash
source ./env

while true; do \
  ./trainbot-arm64; \
done
```

Download latest data from Raspberry Pi:

```bash
ssh "$TRAINBOT_DEPLOY_TARGET_SSH_HOST" sqlite3 data/db.sqlite3
.backup data/db.sqlite3.bak
# Ctrl+D
rsync --verbose --archive --rsh=ssh "$TRAINBOT_DEPLOY_TARGET_SSH_HOST:data/" data/
rm data/db.sqlite3-shm data/db.sqlite3-wal
mv data/db.sqlite3.bak data/db.sqlite3
```

### Web frontend

Images and database are uploaded to a web server via FTP.
The frontend served as a static HTML/JS bundle from the same server.
All database access happens in the browser via sql.js.

## Code notes

* Zerolog is used as logging framework
* "Library" code uses `panic()`, "application" code use `log.Panic()...`

## Hardware setup

My deployment is installed on my balcony in a waterproof case, as seen in the [MagPi Magazine](https://magpi.raspberrypi.com/issues/131).

The case is this one from AliExpress: https://www.aliexpress.com/item/1005003010275396.html

Mounting plate for Camera: https://www.tinkercad.com/things/1FowVwonymJ

Mounting plate for Raspberry Pi: https://www.tinkercad.com/things/djlEF6oQSY1

The prints were ordered from JLCPCB.
Note that the mounting plate for the Raspberry Pi is 1-2mm too wide, because the 86mm stated in the picture on the Aliexpress product page are in reality a bit less. You can solve that by changing the 3d design, or by cutting off a bit from the print. It might however also depend on your specific case.

## TODOs

- [ ] Add machine learning to classify trains (MobileNet, EfficientNet, https://mediapipe-studio.webapps.google.com/demo/image_classifier)
- [ ] Better deployment setup (at least a systemd unit)
- [ ] Add run/deploy instructions to README (including confighelper)
- [ ] Maybe compress URL params - favorites list is getting longer and longer...
