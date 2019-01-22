<template>
    <div class="node">
        <div class="top">
            <span v-if="hasKey()"><span class="key">"{{ keyName }}</span>":</span>
            <span v-if="isArray()">[</span>
            <span v-if="isNode()">{</span>
            <span v-if="isScalar(node)" class="scalar" v-bind:class="getScalarType(node)">
                {{ node | formatScalar }}
                <span v-if="lastElem">,</span>
            </span>
        </div>
        <div class="body">
            <div class="elem" v-if="isNode()" v-for="elem,i in node">
                <preview-json-node v-bind:key-name="i" v-bind:node="elem" v-bind:level="getNextLevel()" v-bind:last-elem="i<node.length-1"></preview-json-node>
            </div>
            <div class="elem" v-if="isArray()" v-for="elem,i in node">
                <preview-json-node v-bind:node="elem" v-bind:level="getNextLevel()" v-bind:last-elem="i<node.length-1"></preview-json-node>
            </div>

            <span v-if="isArray()">]<span v-if="lastElem">,</span></span>
            <span v-if="isNode()">}<span v-if="lastElem">,</span></span>
        </div>
    </div>
</template>

<script>
    define(['Vue'], function (Vue) {
        Vue.component('preview-json-node', {
            template: template,
            props: ['keyName', 'node', 'level', 'lastElem'],
            filters: {
                formatScalar: function (val) {
                    switch (typeof val) {
                        case "string":
                            return '"'+val+'"';

                    }
                    return val;
                }
            },
            methods: {
                getNextLevel: function() {
                    if (!this.level) {
                        return 1;
                    }
                    return parseInt(this.level) + 1;
                },
                getShift: function () {
                    return (this.level * 20) + 'px';
                },
                hasKey: function () {
                    console.log('check key', this.keyName);
                    if ('undefined' === typeof this.keyName) {
                        return false;
                    }
                    return null !== this.keyName;
                },
                isArray: function () {
                    return Array.isArray(this.node)
                },
                isNode: function () {
                    return false === Array.isArray(this.node) && 'object' === typeof this.node && null !== this.node;
                },
                isScalar: function (val) {
                    return null === val || 'object' !== typeof val;
                },
                getScalarType: function (val) {
                    if (val === null) {
                        return "null"
                    }
                    var type = typeof val;
                    switch (type) {
                        case "number":
                            return "number";
                        case "string":
                            return "string";
                        case "boolean":
                            return "bool";
                    }
                }
            }
        });
    })
</script>

<style scoped>
    .node {
        margin-left: 20px;
    }
    .scalar.number {
        color: darkblue;
    }
    .key, .scalar.string {
        color: darkred;
    }
    .scalar.bool {
        color: purple;
    }
    .scalar.null {
        color: red;
    }
</style>