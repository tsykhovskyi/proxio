<template>
    <div class="container-fluid">
        <div class="row">
            <div class="col-lg-6">
                <h4>
                    Requests ({{messages.length}})
                    <button class="btn btn-default pull-right" v-on:click="clear()">Clear</button>
                </h4>
                <table class="table table-striped">
                    <tr v-for="m, i in messages" v-on:click="activate(i, $event)">
                        <td>{{ m.Id }}</td>
                        <td>{{ m.Request.Method }} {{ m.Request.URI }}</td>
                        <td>{{ m.Response && m.Response.Code}}</td>
                        <td>{{ m.Time.TimeTaken}}</td>
                    </tr>
                </table>
            </div>
            <div class="col-lg-6">
                <div v-if="s">
                    <div class="card">
                        <div class="card-header">
                            {{ s.Request.Method }} {{ s.Request.URI }}
                        </div>
                        <div class="card-body">
                            <ul class="nav nav-tabs">
                                <li class="nav-item">
                                    <a class="nav-link" href="#">Pretty</a>
                                </li>
                                <li class="nav-item">
                                    <a class="nav-link active" href="#">Body</a>
                                </li>
                                <li class="nav-item">
                                    <a class="nav-link" href="#">Headers</a>
                                </li>
                            </ul>
                        </div>
                    </div>
                    <div class="card" v-if="s.Response">
                        <div class="card-header">
                            Response
                        </div>
                        <div class="card-body">
                            {{ s.Response.Body }}
                        </div>
                    </div>
                </div>
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
                    s: null // selected message
                }
            },
            created: function () {
                var self = this;

                $http.getJson('/m', function (data) {
                    data.forEach(function (message) {
                        $storage.add(message);
                        self.messages = $storage.getMessages();
                    })
                });

                $http.wsJSON('ws', function (data) {
                    $storage.add(data);
                    self.messages = $storage.getMessages();
                });
            },
            methods: {
                activate: function (i, event) {
                    this.s = this.messages[i]
                },
                clear: function () {
                    let self = this;
                    $http.get('/clear', function () {
                        $storage.removeAll();
                        self.messages = $storage.getMessages();
                    });
                }
            }
        });
    });
</script>
