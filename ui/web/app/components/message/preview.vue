<template>
    <div><h1>Preview</h1></div>
</template>

<script>
    define(['Vue'], function (Vue) {
        Vue.component('message-preview', {
            template: template,
            props: ['content-type', 'body'],
            data: function () {
                return {
                    body: '',
                    resActiveTab: 1,
                };
            },
            methods: {

            }
        });
    })
</script>