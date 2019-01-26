<template>
    <div class="json-preview-box">
        <preview-json-node v-bind:node="json"></preview-json-node>
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
                    console.log('content changed');
                    this.parseContent();
                }
            },
            data: function () {
                return {
                    json: {}
                };
            },
            methods: {
                parseContent: function () {
                    var content = this.content;
                    var json = JSON.parse(content);

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