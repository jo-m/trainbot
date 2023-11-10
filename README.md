# Onlytrains

<img src="frontend/src/assets/logo-day.svg" height="100" width="100">

Watches a piece of train track, detects passing trains, and stitches together images of them.
Should work with any video4linux USB cam, or Raspberry Pi camera v3 modules.

Frontend: <https://trains.jo-m.ch/>

[A collection of some "special" sightings.](https://trains.jo-m.ch/#/trains/list?filter={%22where%22:{%22favs%22:%22id+IN+(577,2405,2320,2193,1342,1039,343,407,350,307,1724,887,2485,3002,2950,2949,2896,2870,2853,2839,2827,2815,2802,3403,3224,3008,2766,2483,3410,3425,3424,3592,3576,3986,3715,2462,3846,3903,3981,3999,3971,4045,4160,4051,4362,4300,4504,4484,4456,4669,4794,4792,4790,4796,4797,4801,4813,4814,4815,4816,4818,4820,4827,4829,4831,4841,4840,4839,4876,4874,4873,4855,4844,4894,4890,4889,4883,4882,5058,5045,5272,5257,5241,5148,5146,4823,5437,2754,3770,3768,4025,4158,4426,4430,5325,5401,6124,6567,6560,6553,5972,5535,5700,6786,7232,8332,8334,8137,7911,8532,8518,8496,8415,7956,7939,7136,7000,7001,9328,9321,9286,9281,9213,9207,9188,9649,9648,9621,9614,9584,9529,9528,9450,9422,9268,9231,9179,9175,8786,8588,11529,11420,11406,11224,11292,11223,11199,11101,11094,11027,11014,10773,10626,10349,10333,9846,9866,9814,12424,12345,12324,12216,12219,12221,12226,12235,12556,13024,12607,13267,13989,13988,13979,13914,13909,13896,13886,13728,13513,13507,13461,12410,12331,14193,14184,14213,14252,14336,14362,14373,13420,3643,13489,13460,13499,15347,15276,15263,15201,15068,15033,14985,14809,14821,14702,15443,15435,15414,15374,17103,17089,17088,17087,17084,17063,17058,17054,16954,16952,16896,16895,16890,16878,16876,16864,16856,16838,16690,16644,16388,16386,16296,16283,16282,16269,16255,16239,16152,16127,16113,16100,16086,16072,16053,15923,15885,15879,15877,15874,15783,15691,15673,15615,15577,15564,18079,14479,138,11810,18090,17983,17974,17962,17956,17860,17581,17536,17472,17471,17468,17305,17252,17211,17190,18288,18599,18538,20417,20421,20150,19604,19515,19259,19260,20426,15457,21289,20885,20862,20818,20808,20479,19387,19342,19317,18342,21282,21215,21098,21085,21782,21775,21749,21737,22835,19805,20863,22033,22317,22360,22655,23107,23215,23185,23244,23253,23258,23373,23445,23510,23538,23587,23562,23614,23671,23905,23906,23923,24016,24046,25260,25205,25087,25054,25024,24930,24508,24496,24487,24409,24378,24274,25636,25617,25616,25595,25587,25533,25298,27421,27388,27305,27301,27157,27154,27141,26997,26923,26811,26752,26713,26701,26588,26535,26521,26412,26394,26281,26274,26198,26141,25851,25777,25776,25752,25664,28510,28505,28393,28475,28503,28390,28310,28241,28180,28080,27560,29291,29282,29275,29252,29158,29096,29088,32739,32732,32702,32694,32590,32395,32380,33881,33855,33850,33816,33803,33725,33675,33489,33390,33352,33320,33309,33290,33165,33070,33063,33057,33044)%22}})

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

How to get binaries?
There are multiple options:

1. `go install github.com/jo-m/trainbot/cmd/trainbot@latest` - Let Go build the binary for your host system.
2. Grab a binary from the latest CI run at https://github.com/jo-m/trainbot/actions
3. Use the Docker setup in the repo. This will build for Linux `x86_64` and `arm64`:

```bash
git clone https://github.com/jo-m/trainbot
cd trainbot
make docker_build

# Find binaries in build/ after this has completed.
```

### Raspberry Pi

```bash
sudo usermod -a -G video pi

# confighelper
./confighelper-arm64 --log-pretty --input=picam3 --listen-addr=0.0.0.0:8080
```

The current production deployment is in a Tmux session... to be improved one day, but it has worked for 6 months now.

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
