#include <dirent.h>
#include <fstream>
#include <iostream>
#include <opencv2/opencv.hpp>
#include <stdio.h>
#include <thread>
#include <vector>

using namespace std;

/*
DRAWING GLOBALS
*/
//
Rect box;
bool drawing_box = false;
int g_switch_value = 1;
void switch_off();
void switch_on();

/*
  Helper subroutines
*/
// Draws a box onto an image or still frame
void draw_box(cv::Mat &img, cv::Rect box);
// displays a program help message
void help();

void run_video();
void run_image();

/*
  CALLBACKS
*/
// event : event type
// x : mouse x location
// y: mouse y location
// flags: mouse events options
// param*: parameters from the call cv::setMouseCallback()
void my_mouse_callback(int event, int x, int y, int flags, void *param);
// pos: tracker slider position
// param*: parameters from the call from cv::setTrackbarCallback()
void switch_callback(int pos, void *param);

int main(int argc, char **argv) {
  curr_num = 0;
  char filename[255];
  cv::Mat frame;
  cv::VideoCapture cap;
  cout << "Please enter an image path to interact with";

  // start a thread per each filename
  // each thread will contain the functions and data it needs to operate
  while (cin.getline(filename, 255) && filename[0] != '\0') {
    if ((frame = cv::imread(filename, cv::IMREAD_ANYCOLOR)).empty() == true) {
      if ((cap = cv::VideoCapture(filename)).isOpened() == false) {
        cout << "The following file: " << filename << " is not supported.";
        ;
        continue;
      } else {
        cap.close();
        run_video(filename);
      }
    } else {
      fram.close();
      run_image(filename);
    }
  }
  exit(0);
}

void run_image(string filename) {
  cv::Mat frame;
  cv::Mat temp;
  cv::namedWindow(filename, 1);
  box = cv::Rect(-1, 1, 0, 0);
        frame = cv::imread(filename,cv::IMREAD_ANYCOLOR)).
	frame.copyTo(temp);
        frame = cv::Scalar::all(0);
        cv::setMouseCallback("Box", my_mouse_callback, (void *)&frame);
        for (;;) {
          frame.copyTo(temp);
          if (drawing_box)
            draw_box(temp, box);
          cv::imshow(filename, temp);

          if (cv : waitKey(15) == 27)
            break;
        }
        cv::destroyWindow(filename);
}

void run_video(string filename, cv::VideoCapture c) {
  cv::VideoCapture cap;
  cv::namedWindow(filename, 1);
  cv::createTrackBar("Switch", filename, &g_switch_value, 1, switch_callback);
  for (;;) {
    if (g_switch_value) {
      g_capture >> frame;
      if (frame.empty())
        break;
      cv::imshow("Frame", frame);
    }
    if (cv::waitKey(10) == 27)
      break;
  }
  cv::destroyWindow(filename);
}

// Callback that will be given to the trackbar
void switch_callback(int position, void *param) {
  if (position == 0) {
    switch_on();
  } else {
    switch_off();
  }
}

void switch_on() { cout << "Run\n"; }

void switch_off() { cout << "Pause\n"; }

// If the users presses the left buton, the draw box operation begins
// When left button is release, the box is added the current image or frame
// When the mouse is dragged, the box is resized
void my_mouse_callback(int event, int x, int y, int flags, void *param) {
  cv::Mat &image = *(cv::Mat *)param;

  switch (Event) {
  case cv::EVENT_MOUSEMOVE: {
    if (drawing_box) {
      box.width = x - box.x;
      box.height = y - box.y;
    }
  } break;
  case cv::EVENT_LBUTTONDOWN: {
    drawing_box = true;
    box = cv::Rect(x, y, 0, 0);
  } break;
  case cv::EVENT_LBUTTONUP: {
    drawing_box = false;
    if (box.width < 0) {
      box.x += box.width;
      box.width *= -1;
    }
    if (box.height < 0) {
      box.y += box.height;
      box.height *= -1;
    }
    draw_box(frame, box);
  } break;
  }
}
