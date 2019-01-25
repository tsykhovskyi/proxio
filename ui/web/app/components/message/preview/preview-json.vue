<template>
    <div class="json-preview-box">
        <preview-json-node v-bind:node="json" v-bind:hovered-blocks="[]"></preview-json-node>
    </div>
</template>

<script>
    define(['Vue'], function (Vue) {
        Vue.component('preview-json', {
            template: template,
            props: ['content'],
            created: function() {
                this.parseContent();
            },
            watch: {
                content: function () {
                    this.parseContent();
                }
            },
            data: function () {
                return {
                    data: null
                };
            },
            methods: {
                parseContent: function () {
                    var content = this.content;
                    var json = JSON.parse(content);
                    console.log(json);

                    this.json = json;
                },
            }
        });
    })
</script>

<style scoped>
    .json-preview-box > div {
        margin: 0 !important;
    }
</style>