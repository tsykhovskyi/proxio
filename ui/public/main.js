var http = (function () {
    function sendRequest(method, url, cbReadyState, dataType) {
        dataType = 'undefined' === typeof dataType ? 'text' : dataType;

        var xhr = new XMLHttpRequest();
        xhr.open(method, url, true);
        xhr.responseType = dataType;
        xhr.onreadystatechange = function () {
            if (xhr.readyState == XMLHttpRequest.DONE) {
                cbReadyState(xhr);
            }
        };
        xhr.send();
    }

    function get(url, cbSuccess) {
        sendRequest('GET', url, function(xhr) {
            cbSuccess(xhr.response);
        });
    }

    function getJson(url, cbSuccess, cbError) {
        sendRequest('GET', url, function(xhr) {
            cbSuccess(xhr.response);
        }, 'json');
    }

    function getJsonRecursively(url, cbSuccess, cbError) {
        getJson(url, function (data) {
            cbSuccess(data);

            getJsonRecursively(url, cbSuccess, cbError);
        }, cbError)
    }

    return {
        get: get,
        getJson: getJson,
        getJsonRecursively: getJsonRecursively
    }
}());

var storage = (function () {
    var messages = [];

    function add(m) {
        for (i in messages) {
            if (messages[i].Id === m.Id && m.Response !== null) {
                messages[i] = m;
                return
            }
        }
        messages.push(m);
    }

    function sortMessagesByCreated() {
        console.log("sorted");
        messages = messages.sort(function (a, b) {
            if (a.Time.StartedAt < b.Time.StartedAt) {
                return 1
            }
            if (a.Time.StartedAt > b.Time.StartedAt) {
                return -1
            }
            return 0
        })
    }

    function removeAll() {
        messages = []
    }

    function getMessages() {
        sortMessagesByCreated();
        return messages;
    }

    return {
        getMessages: getMessages,
        add: add,
        removeAll: removeAll
    }
})();

(function (http, storage) {
    var proxio = new Vue({
        el: '#proxio',
        data: {
            messages: [],
            s: null // selected message
        },
        methods: {
            activate: function (i, event) {
                this.s = this.messages[i]
            },
            clear: function () {
                let self = this;
                http.get('/clear', function () {
                    storage.removeAll();
                    self.messages = storage.getMessages();
                    self.$forceUpdate();
                });
            }
        }
    });

    http.getJson('/m', function (data) {
        data.forEach(function (message) {
            storage.add(message);
            proxio.messages = storage.getMessages();
            proxio.$forceUpdate();
        })
    });

    http.getJsonRecursively("/check", function (data) {
        data.forEach(function (message) {
            storage.add(message);
            proxio.messages = storage.getMessages();
            proxio.$forceUpdate();
        })
    });
})(http, storage);