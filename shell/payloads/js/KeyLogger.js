var $_$timer = null,
	$_$buffer = [],
	$_$self = this;
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