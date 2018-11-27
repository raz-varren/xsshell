var $_$self = this,
    $_$createReq = function() {
        var XMLHttpFactories = [
                function () {return new XMLHttpRequest()},
                function () {return new ActiveXObject("Msxml3.XMLHTTP")},
                function () {return new ActiveXObject("Msxml2.XMLHTTP.6.0")},
                function () {return new ActiveXObject("Msxml2.XMLHTTP.3.0")},
                function () {return new ActiveXObject("Msxml2.XMLHTTP")},
                function () {return new ActiveXObject("Microsoft.XMLHTTP")}
            ],
            xmlhttp = false;

        for (var i=0;i<XMLHttpFactories.length;i++) {
            try {
                xmlhttp = XMLHttpFactories[i]();
            }catch (e) { continue; }
            break;
        }
        return xmlhttp;
    },
    $_$sendRequest = function(url, contentHeader, postData) {
        var req = $_$createReq();
        if (!req) return;
        var method = (postData) ? "POST" : "GET";
        req.open(method,url,true);
        if (postData){
            req.setRequestHeader('Content-type', contentHeader);
        }
        req.onreadystatechange = function () {
            if (req.readyState != 4) return;
            if (req.status != 200 && req.status != 304) {
                $_$self.send("xhr request failed, status: "+req.status+"\nresponse: "+req.responseText);
                return;
            }
            $_$self.send(req.responseText);
        }
        if (req.readyState == 4) return;
        req.send(postData);
    };

