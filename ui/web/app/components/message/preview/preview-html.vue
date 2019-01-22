<template>
    <iframe v-bind:srcdoc="content" frameborder="0"></iframe>
</template>

<script>
    define(['Vue'], function (Vue) {
        Vue.component('preview-html', {
            template: template,
            props: ['content'],
        });
    })
</script>

<style scoped>
    iframe {
        border: 1px dotted #333;
        width: 100%;
        height: 500px;
    }
</style>