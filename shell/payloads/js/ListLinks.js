var $_$links = document.querySelectorAll('a'),
	$_$buffer = [];

for(var i=0; $_$links != null && i<$_$links.length; i++){
	var $_$link = $_$links[i];
	$_$buffer.push({
		text: $_$link.innerText,
		href: $_$link.href
	});
}

this.send(JSON.stringify($_$buffer));