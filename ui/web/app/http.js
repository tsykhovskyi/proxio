define(function () {
    var requestsIdentifier = null,
        REQUEST_ID_HEADER = 'Requests-Identifier'
    ;

    function sendRequest(method, url, cbReadyState, dataType) {
        dataType = 'undefined' === typeof dataType ? 'text' : dataType;

        var xhr = new XMLHttpRequest();
        xhr.open(method, url, true);
        if (null !== requestsIdentifier) {
            xhr.setRequestHeader(REQUEST_ID_HEADER, requestsIdentifier);
        }
        xhr.responseType = dataType;
        xhr.onreadystatechange = function () {
            if (this.readyState == this.HEADERS_RECEIVED) {
                var id = xhr.getResponseHeader(REQUEST_ID_HEADER);
                if (null !== id) {
                    requestsIdentifier = id;
                }
            }
            if (xhr.readyState == XMLHttpRequest.DONE) {
                cbReadyState(xhr);
            }
        };
        xhr.send();
    }

    function get(url, cbSuccess) {
        sendRequest('GET', url, function (xhr) {
            cbSuccess(xhr.response);
        });
    }

    function getJson(url, cbSuccess, cbError) {
        sendRequest('GET', url, function (xhr) {
            cbSuccess(xhr.response);
        }, 'json');
    }

    function getJsonRecursively(url, cbSuccess, cbError) {
        getJson(url, function (data) {
            cbSuccess(data);

            getJsonRecursively(url, cbSuccess, cbError);
        }, cbError)
    }

    function ws(uri, cbFrameRecieved) {
        var loc = window.location, new_uri;
        new_uri = "ws://" + loc.host + loc.pathname + uri;
        var socket = new WebSocket(new_uri);
        // Connection opened
        socket.addEventListener('open', function (event) {
            // socket.send(event);
        });

        // Listen for messages
        socket.addEventListener('message', function (event) {
            cbFrameRecieved(event)
        });

        return socket
    }

    function wsJSON(uri, cbDataRecieved) {
        return ws(uri, function (event) {
            cbDataRecieved(JSON.parse(event.data));
        })
    }

    return {
        get: get,
        getJson: getJson,
        wsJSON: wsJSON,
    }
});