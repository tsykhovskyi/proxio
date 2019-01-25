<template>
    <div class="node" @mouseenter="mouseEntered()" @mouseleave="mouseLeft()" v-bind:class="{hovered: isHover}">
        <div class="top">
            <span v-if="hasKey()"><span class="property">"{{ propertyName }}</span>":</span>
            <span v-if="isArray()">[</span>
            <span v-if="isNode()">{</span>
            <span v-if="isScalar(node)" class="scalar" v-bind:class="getScalarType(node)">
                {{ node | formatScalar }}<span v-if="lastProp === false">,</span>
            </span>
        </div>
        <div class="node-struct" v-if="isNode() || isArray()" >
            <div class="elem" v-for="(elem, index) in node">
                <preview-json-node  v-bind:property-name="index" v-bind:node="elem" v-bind:level="getNextLevel()" v-bind:last-prop="index === getLastKey(node)"></preview-json-node>
            </div>

            <span v-if="isArray()">]</span>
            <span v-if="isNode()">}</span>
            <span v-if="lastProp === false">,</span>
        </div>
    </div>
</template>

<script>
    define(['Vue'], function (Vue) {
        Vue.component('preview-json-node', {
            template: template,
            props: ['propertyName', 'node', 'level', 'lastProp', 'hovered-blocks'],
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
                    isHover: false
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
                getLastKey: function(obj) {
                    if (Array.isArray(obj)) {
                        return obj.length - 1;
                    }
                    var keys = Object.keys(obj);
                    return keys[keys.length - 1];
                },
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
                    if ("number" === typeof this.propertyName) {
                        return false;
                    }
                    if ('undefined' === typeof this.propertyName) {
                        return false;
                    }
                    return null !== this.propertyName;
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
        margin-left: 30px;
        transition: background-color 0.3s ease-in;
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
    .hovered {
        background-color: rgba(235, 238, 249, 1);
    }
</style>