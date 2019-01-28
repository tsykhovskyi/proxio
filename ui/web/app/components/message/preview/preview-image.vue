<template>
    <div class="image">
        <h3>Image</h3>
        <img :src="src">
    </div>
</template>

<script>
    define(['Vue'], function (Vue) {
        Vue.component('preview-image', {
            template: template,
            props: ['content'],
            created: function () {
                this.src = this.prepareSrc(this.content);
            },
            watch: {
                content: function (val) {
                    this.src = this.prepareSrc(val);
                }
            },
            data: function () {
                return {
                    src: "https://interactive-examples.mdn.mozilla.net/media/examples/grapefruit-slice-332-332.jpg"
                };
            },
            methods: {
                prepareSrc: function (content) {
                    return "data:image/jpeg;base64," + content;
                }
            }
        });
    })
</script>

<style scoped>

</style>