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

require(["Vue", "vue!components/proxio"], function(Vue){
    new Vue({
        el: "#app",
        template: "<proxio></proxio>"
    });
});