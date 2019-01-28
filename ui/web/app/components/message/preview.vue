<template>
    <div>
        <preview-html v-if="contentType === 0" v-bind:content="body"></preview-html>
        <preview-json v-if="contentType === 1" v-bind:content="body"></preview-json>
        <preview-image v-if="contentType === 2" v-bind:content="body"></preview-image>
    </div>

</template>

<script>
    define(['Vue'], function (Vue) {
        Vue.component('message-preview', {
            template: template,
            props: ['headers', 'body', 'i'],
            created: function () {
                this.detectContent();
            },
            data: function () {
                return {
                    // 0 - html
                    // 1 - json
                    // 2 - image
                    contentType: null
                };
            },
            watch: {
                headers: function () {
                    this.detectContent()
                }
            },
            methods: {
                detectContent: function () {
                    if ('object' === typeof this.headers['Content-Type']) {
                        var type = this.headers['Content-Type'][0],
                            rHtml = /text\/html/,
                            rJson = /application\/json/,
                            rImage = /image\/.+/
                        ;
                        if (rHtml.test(type)) {
                            this.contentType = 0;
                            return
                        }
                        if (rJson.test(type)) {
                            this.contentType = 1;
                            return
                        }
                        if (rImage.test(type)) {
                            this.contentType = 2;
                            return;
                        }
                    }
                    this.contentType = null
                }
            }
        });
    })
</script>