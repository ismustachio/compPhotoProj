Computational Photography Project
This project attempts to act as a command-line utility, image filter,
application. The program has the following avaliable image filter:
	LeftSobel
	RightSobel
	TopSobel
	BottomSobel
	Gaussian
	Emboss
	Identity
	Outline
	Sharp
	Blur
	Custom (3x3)
To maxize this utility, it is best to run against
large quantaties of images. This utility will spawned a thread
(goroutine) per each image using all available system resources.


Example Usage:
  ./compPhoto -p image.png 	-f Emboss -r 3
  ./compPhoto -p ./testImages/ -f Identity
  ./compPhoto -p image1.png -c '.3 .2 -1 0 1 1 .8 .2 .2' -r 10


