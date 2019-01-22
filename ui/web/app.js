requirejs.config({
    baseUrl: 'lib',
    paths: {
        app: '../app',
        components: '../app/components',

        Vue: './vue',
        vue: "./requirejs-vue"
    },
    shim: {
        "Vue": {"exports": "Vue"}
    }
});

require([
    "Vue",
    "vue!components/proxio",
    "vue!components/message",
    "vue!components/message/headers",
    "vue!components/message/preview",
    "vue!components/message/preview/preview-html",
    "vue!components/message/preview/preview-json",
    "vue!components/message/preview/preview-json-node"
], function (Vue) {
    new Vue({
        el: "#app",
        template: "<proxio></proxio>"
    });
});