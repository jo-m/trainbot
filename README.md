# Trainbot

**THIS IS A WORK IN PROGRESS AND INCOMPLETE**

Trainbot watches a piece of train track with a USB camera, detects trains, and stitches together images of them.

[<img src="pkg/stitch/testdata/test1.jpg">](pkg/stitch/testdata/test1.jpg)
[<img src="pkg/stitch/testdata/test2.jpg">](pkg/stitch/testdata/test2.jpg)
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

## Code notes

* Zerolog is used as logging framework
* "Library" code uses `panic()`, "application" code use `log.Panic()...`

## TODOs

- [x] Also create GIFs
- [ ] Crop stitched images to exact width and height
- [ ] Use FFMPEG or Gstreamer for camera input, the Go webcam library often crashes after a couple 100s of frames
- [ ] Use https://github.com/stapelberg/turbojpeg for faster jpeg encoding on output
- [ ] Move all "application" code to internal/
- [ ] Deploy to Raspberry Pi via [gokrazy](https://gokrazy.org/)
- [ ] Maybe move patchmatch to separate repo
- [ ] Add run/deploy instructions to README (including confighelper)
- [ ] Add Telegram or Twitter bot, or serve a page with recent trains
- [ ] Improve stiching seams
- [ ] Document build/cross-build
- [ ] Swap out GIF palletor for something which allows to set a random seed, so it can have deterministic tests
- [ ] Measure FPS on host and RasPi 4
- [ ] Maybe combine with https://github.com/jo-m/gocatprint to directly print trains on paper
- [ ] Stop relying on FPS
- [ ] Unified frame source factory (device vs. video)
