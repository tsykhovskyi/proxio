define(function () {
    var messages = [];

    function add(m) {
        for (i in messages) {
            if (messages[i].Id === m.Id) {
                messages[i] = m;
                return
            }
        }
        messages.push(m);
    }

    function sortMessagesByCreated() {
        messages = messages.sort(function (a, b) {
            if (a.Id < b.Id) {
                return 1
            }
            if (a.Id > b.Id) {
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
});