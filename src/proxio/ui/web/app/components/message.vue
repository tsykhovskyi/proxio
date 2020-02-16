<template>
    <div v-if="message">
        <div class="card">
            <div class="card-header">
                {{ message.Request.Method }} {{ message.Request.URI }}
            </div>
            <div class="card-body">
                <ul class="nav nav-tabs">
                    <li class="nav-item">
                        <a class="nav-link" v-bind:class="{'active': isReqTab(1)}" v-on:click="setReqTab(1)" href="#">Pretty</a>
                    </li>
                    <li class="nav-item">
                        <a class="nav-link" v-bind:class="{'active': isReqTab(2)}" v-on:click="setReqTab(2)" href="#">Body</a>
                    </li>
                    <li class="nav-item">
                        <a class="nav-link" v-bind:class="{'active': isReqTab(3)}" v-on:click="setReqTab(3)" href="#">Headers</a>
                    </li>
                </ul>
                <div class="tab-content">
                    <div class="tab-pane" v-bind:class="{'active': isReqTab(1)}">
                        <message-preview
                                v-bind:headers="message.Request.Headers"
                                v-bind:body="message.Request.Body"
                                v-bind:i="message.Id"
                        ></message-preview>
                    </div>
                    <div class="tab-pane" v-bind:class="{'active': isReqTab(2)}">
                        <pre>{{ message.Request.Body }}</pre>
                    </div>
                    <div class="tab-pane" v-bind:class="{'active': isReqTab(3)}">
                        <message-headers
                                v-if="message.Request.Headers"
                                v-bind:headers="message.Request.Headers"
                        ></message-headers>
                    </div>
                </div>
            </div>
        </div>
        <div class="card" v-if="message.Response">
            <div class="card-header">
                {{ message.Response.Code }}
            </div>
            <div class="card-body">
                <ul class="nav nav-tabs">
                    <li class="nav-item">
                        <a class="nav-link" v-bind:class="{'active': isResTab(1)}" v-on:click="setResTab(1)" href="#">Pretty</a>
                    </li>
                    <li class="nav-item">
                        <a class="nav-link" v-bind:class="{'active': isResTab(2)}" v-on:click="setResTab(2)" href="#">Body</a>
                    </li>
                    <li class="nav-item">
                        <a class="nav-link" v-bind:class="{'active': isResTab(3)}" v-on:click="setResTab(3)" href="#">Headers</a>
                    </li>
                </ul>
                <div class="tab-content">
                    <div class="tab-pane" v-bind:class="{'active': isResTab(1)}">
                        <message-preview
                                v-bind:headers="message.Response.Headers"
                                v-bind:body="message.Response.Body"
                        ></message-preview>
                    </div>
                    <div class="tab-pane" v-bind:class="{'active': isResTab(2)}">
                        <pre>{{ message.Response.Body }}</pre>
                    </div>
                    <div class="tab-pane" v-bind:class="{'active': isResTab(3)}">
                        <message-headers v-if="message.Response.Headers"
                                         v-bind:headers="message.Response.Headers"></message-headers>
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>

<script>
    define(['Vue'], function (Vue) {
        Vue.component('message', {
            template: template,
            props: ['message'],
            data: function () {
                return {
                    reqActiveTab: 1,
                    resActiveTab: 1,
                };
            },
            methods: {
                isReqTab: function (name) {
                    return this.reqActiveTab === name;
                },
                setReqTab: function (name) {
                    this.reqActiveTab = this.reqActiveTab === name ? null : name;
                },
                isResTab: function (name) {
                    return this.resActiveTab === name;
                },
                setResTab: function (name) {
                    this.resActiveTab = this.resActiveTab === name ? null : name;
                },
            }
        });
    })
</script>