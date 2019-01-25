<template>
    <div class="node" @mouseenter="mouseEntered()" @mouseleave="mouseLeft()" v-bind:class="{hovered: isHover}">
        <span class="collapser" v-if="isNode()" v-on:click="collapse()">{{ isCollapsed?'-':'+' }}</span>
        <div class="top">
            <span v-if="hasKey()"><span class="property">"{{ propertyName }}</span>":</span>

            <span v-if="isArray()">[</span>
            <span v-if="isObject()">{</span>
            <span v-if="isCollapsed" class="collapsed-block"><span v-on:click="collapse()">...</span> {{ getClosedString() }}</span>

            <span v-if="isScalar(node)" class="scalar" v-bind:class="getScalarType(node)">
                {{ node | formatScalar }}<span v-if="lastProp === false">,</span>
            </span>
        </div>
        <div class="node-struct" v-if="isNode()" v-bind:class="{hidden: isCollapsed}">
            <div class="elem" v-for="(elem, index) in node">
                <preview-json-node  v-bind:property-name="index" v-bind:node="elem" v-bind:level="getNextLevel()" v-bind:last-prop="index === getLastKey(node)"></preview-json-node>
            </div>
            <span>{{ getClosedString() }}</span>
        </div>
    </div>
</template>

<script>
    define(['Vue'], function (Vue) {
        Vue.component('preview-json-node', {
            template: template,
            props: ['propertyName', 'node', 'level', 'lastProp'],
            filters: {
                formatScalar: function (val) {
                    switch (typeof val) {
                        case "string":
                            return '"'+val+'"';
                        case "object":
                            if (val === null) {
                                return "null";
                            }
                    }
                    return val;
                }
            },
            created: function () {
                if('undefined' === typeof Vue.$selectedJsonBlocks) {
                    Vue.$selectedJsonBlocks = [];
                }
                this.selectedBlocks = Vue.$selectedJsonBlocks;
            },
            data: function () {
                return {
                    selectedBlocks: null,
                    isHover: false,
                    isCollapsed: false
                }
            },
            watch: {
                selectedBlocks: function (blocks) {
                    var lastElem = blocks[blocks.length - 1];
                    this.isHover = lastElem === this._uid;
                }
            },
            methods: {
                mouseEntered: function () {
                    this.selectedBlocks.push(this._uid);
                },
                mouseLeft: function () {
                    this.selectedBlocks.pop();
                },
                collapse: function () {
                    this.isCollapsed = !this.isCollapsed;
                    console.log('coll', this.isCollapsed);
                },
                getLastKey: function(obj) {
                    if (Array.isArray(obj)) {
                        return obj.length - 1;
                    }
                    var keys = Object.keys(obj);
                    return keys[keys.length - 1];
                },
                getClosedString: function (obj) {
                    var str = "";
                    if (this.isArray()) {
                        str = "]"
                    } else if (this.isObject()) {
                        str = "}"
                    } else {
                        return "";
                    }
                    if (this.lastProp === false) {
                        str += ',';
                    }
                    return str;
                },
                getNextLevel: function() {
                    if (!this.level) {
                        return 1;
                    }
                    return parseInt(this.level) + 1;
                },
                hasKey: function () {
                    if ("number" === typeof this.propertyName) {
                        return false;
                    }
                    if ('undefined' === typeof this.propertyName) {
                        return false;
                    }
                    return null !== this.propertyName;
                },
                isNode: function () {
                    return this.isArray() || this.isObject();
                },
                isArray: function () {
                    return Array.isArray(this.node)
                },
                isObject: function () {
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
        margin-left: 30px;
        transition: background-color 0.3s ease-in;
        transition-delay: 0.3s;
    }
    .scalar.number {
        color: darkblue;
    }
    .property, .scalar.string {
        color: darkred;
    }
    .scalar.bool {
        color: purple;
    }
    .scalar.null {
        color: red;
    }
    .node.hovered {
        background-color: rgba(235, 238, 249, 1);
    }
    .collapser {
        position: absolute;
        margin-left: -1em;
        cursor: pointer;
    }
    .hidden {
        display: none;
    }
    .collapsed-block {
        cursor: pointer;
    }
</style>