const url = process.env.URL + "/compress";
// dec2hex :: Integer -> String
function dec2hex (dec) {
    return ('0' + dec.toString(16)).substr(-2)
}
  
// generateId :: Integer -> String
function generateId (len) {
    var arr = new Uint8Array((len || 40) / 2)
    window.crypto.getRandomValues(arr)
    return Array.from(arr, dec2hex).join('')
}

const username = process.env.MQUNAME;
// const username = "1406568753";

const password = process.env.MQPW;
// const password = "167664";

const vhost = process.env.MQVHOST;
// const vhost = "1406568753";

const host = window.location.href.replace("http://", '');

const url = 'http://localhost:20608/compress';

var id = ""

const form = document.querySelector('form');

form.addEventListener('submit', e => {
    e.preventDefault();

    id = generateId()
    console.log(id)
    const files = document.querySelector('[type=file]').files[0];
    const formData = new FormData();
    formData.append('file', files);
    WebSocketTest();
    fetch(url, {
        method: 'POST',
        headers: {
            "X-ROUTING-KEY": id,
        },
        body: formData,
    }).then(function(response) {
        return response.json();
    }).then(function(json) {
        WebSocketTest();
        console.log(JSON.stringify(json));
    });
});

function WebSocketTest() {
	if ("WebSocket" in window) {
		// var ws_stomp_display = new SockJS(process.env.MQURL);
		var ws_stomp_display = new SockJS('http://152.118.148.103:15674/stomp');
		var client_display = Stomp.over(ws_stomp_display);
		// var mq_queue_display = "/exchange/"+ process.env.NPM + "/" + id;
		var mq_queue_display = "/exchange/"+ "1406568753" + "/" + id;
		var on_connect_display = function() {
			console.log('connected');
			client_display.subscribe(mq_queue_display, on_message_display);
		};
		var on_error_display = function() {
			console.log('error');
		};
		var on_message_display = function(m) {
			console.log('message received');
			document.getElementById("status").innerHTML = m.body;
		};
		client_display.connect(username, password, on_connect_display, on_error_display, vhost);
	} else {
		// The browser doesn't support WebSocket
		alert("WebSocket NOT supported by your Browser!");
	}
}