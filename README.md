# Trainbot

**THIS IS A WORK IN PROGRESS AND INCOMPLETE**

Trainbot watches a piece of train track with a USB camera, detects trains, and stitches together images of them.

[<img src="internal/pkg/stitch/testdata/test1.jpg">](internal/pkg/stitch/testdata/test1.jpg)
[<img src="internal/pkg/stitch/testdata/test2.jpg">](internal/pkg/stitch/testdata/test2.jpg)
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

As a consequence, there are certain requirements which have to be met:

1. Trains only appear in a (manually) pre-cropped region.
1. The camera is stable and the image does not move around in any direction.
1. There are no large fast brightness changes.
1. Trains have a given min and max speed.
1. We are looking at the tracks more or less perpendicularly in the chosen image crop region.
1. Trains are coming from one direction at a time, crossings are not yet handled properly.
1. Trains have a constant acceleration (might be 0) and do not stop and turn around while in front of the camera.

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
```

## Running on Raspberry Pi

```bash
./confighelper-arm64 --log-pretty --input=picam3 --listen-addr=0.0.0.0:8080
```

## Code notes

* Zerolog is used as logging framework
* "Library" code uses `panic()`, "application" code use `log.Panic()...`

## TODOs

- [x] Also create GIFs
- [ ] Crop stitched images to exact width and height
- [ ] Use https://github.com/stapelberg/turbojpeg for faster jpeg encoding on output
- [ ] Test in snow/bad weather
- [ ] Deploy to Raspberry Pi via [gokrazy](https://gokrazy.org/)
- [ ] Maybe move patchmatch to separate repo
- [ ] Add run/deploy instructions to README (including confighelper)
- [ ] Add Telegram or Twitter bot, or serve a page with recent trains
- [ ] Improve stiching seams
- [ ] Clean up and document build system
- [ ] Swap out GIF palletor for something which allows to set a random seed, so it can have deterministic tests
- [ ] Measure FPS on host and RasPi 4
- [ ] Maybe combine with https://github.com/jo-m/gocatprint to directly print trains on paper
- [ ] PiCam3Src: list/probe cameras
- [ ] Reconsider failedFrames
- [ ] Maybe make RaspiCam3 sensor mode configurable
