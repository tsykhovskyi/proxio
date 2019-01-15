define(function () {
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
});