<template>
    <div class="container-fluid">
        <div class="row">
            <div class="col-lg-5">
                <div class="row">
                    <div class="col">
                        <h4>
                            Requests ({{messages.length}})
                        </h4>
                    </div>
                    <div class="col text-right">
                        <button class="btn btn-primary" v-on:click="clear()">Clear</button>
                    </div>
                </div>
                <table class="table table-hover requests">
                    <tbody>
                    <tr v-for="(m,i) in messages"
                        v-on:click="activate(i, $event)"
                        v-bind:class="{'table-active': isActive(i), 'table-danger': isMessageCancel(m)}"
                    >
                        <td><span class="badge" v-bind:class="getMethodClass(m)">{{ m.Request.Method }}</span></td>
                        <td class="priority">{{ m.Request.URI }}</td>
                        <td>{{ m.Response && m.Response.Code}}</td>
                        <td>
                            <span v-if="m.Time.TimeTaken > 0">{{ m.Time.TimeTaken | readableTime}}</span>
                        </td>
                    </tr>
                    </tbody>
                </table>
            </div>
            <div class="col-lg-7">
                <message v-bind:message="selected"></message>
            </div>
        </div>
    </div>
</template>

<script>
    define(['app/http', 'app/storage', 'Vue'], function ($http, $storage, Vue) {
        Vue.component('proxio', {
            template: template,
            data: function () {
                return {
                    messages: $storage.getMessages(),
                    selected: null // selected message
                }
            },
            created: function () {
                var self = this;

                $http.getJson('/m', function (data) {
                    data.forEach(function (message) {
                        console.log(message.Id);
                        // message.Response.Body = btoa(message.Response.Body);
                        $storage.add(message);
                        self.messages = $storage.getMessages();
                        self.selected = self.messages.count === 0 ? null : self.messages[0];
                    })
                });

                $http.wsJSON('ws', function (data) {
                    $storage.add(data);
                    self.messages = $storage.getMessages();
                });
            },
            methods: {
                getMethodClass: function (message) {
                    var method = message.Request.Method.toLowerCase();
                    switch (method) {
                        case "get":
                            return "badge-success";
                        case "delete":
                            return "badge-danger";
                        default:
                            return "badge-warning";
                    }
                },
                activate: function (i) {
                    this.selected = this.messages[i]
                },
                isActive: function(i) {
                    return this.selected === this.messages[i]
                },
                isMessageCancel: function (message) {
                    return 1 === message.Status;
                },
                clear: function () {
                    let self = this;
                    $http.get('/clear', function () {
                        $storage.removeAll();
                        self.messages = $storage.getMessages();
                    });
                }
            },
            filters: {
                readableTime: function (value) {
                    var result;
                    if (value < 1) {
                        result = (value * 1000).toFixed(2) + 'ms';
                    } else {
                        result = value.toFixed(2) + 's';
                    }
                    return result;
                }
            }
        });
    });
</script>

<style scoped>
    .priority {
        width: 100%;
    }
    .requests tr {
        cursor: pointer;
    }
</style>
