var http = (function () {
    function getJson(url, cbSuccess, cbError) {
        var xhr = new XMLHttpRequest();
        xhr.open('GET', url, true);
        xhr.responseType = 'json';
        xhr.onreadystatechange = function () {
            if (xhr.readyState == XMLHttpRequest.DONE) {
                console.log(xhr.responseType);
                if (xhr.responseType === 'json') {
                    cbSuccess(xhr.response);
                    delete xhr;
                }
            }
        };
        xhr.send();
    }

    function getJsonRecursively(url, cbSuccess, cbError) {
        getJson(url, function (data) {
            console.log("recursive call");
            cbSuccess(data);

            getJsonRecursively(url, cbSuccess, cbError);
        }, cbError)
    }

    return {
        getJson: getJson,
        getJsonRecursively: getJsonRecursively
    }
}());

var storage = (function () {
    var messages = [];

    function add(m) {
        messages.push(m);
    }

    return {
        messages: messages,
        add: add
    }
})();

(function (http, storage) {
    http.getJson('/m', function (data) {
        data.forEach(function (message) {
            storage.add(message)
        })
    });

    http.getJsonRecursively("/check", function (data) {
        data.forEach(function (message) {
            storage.add(message)
        })
    });

    var reqBox = new Vue({
        el: '#reqBox',
        data: {
            requests: storage.messages
        }
    });
})(http, storage);