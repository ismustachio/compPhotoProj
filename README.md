<h1>
Computational Photography Project
</h1>
<p>
This project attempts to act as a command-line utility, image filter. The program has the following avaliable image filter:
	<ol>
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
		</ol>
To maxize this utility, it is best to run against
large batch of images. This utility will spawned a thread
(goroutine) per each image. The more resources available the 
faster the computation should be. This project benefits from
being embarrassingly parallel, and should scale linearly as the
amount of parallelism increases.
</p>

<p>
Example Usage:
	<ol>
  ./compPhoto -p image.png 	-f Emboss -r 3
		</ol>
		<ol>
  ./compPhoto -p ./testImages/ -f Identity
	</ol>
	<ol>
  ./compPhoto -p image1.png -c '.3 .2 -1 0 1 1 .8 .2 .2' -r 10
  </ol>
</p>

