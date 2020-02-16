<template>
    <table class="table table-striped">
        <tbody>
        <tr v-for="data, name in headers">
            <td>{{ name }}</td>
            <td>
                <span v-for="val in data">{{ val }}</span>
            </td>
        </tr>
        </tbody>
    </table>
</template>

<script>
    define(['Vue'], function (Vue) {
        Vue.component('message-headers', {
            template: template,
            props: ['headers']
        });
    })
</script>

<style scoped>
    table tr td:nth-child(2) {
        word-wrap: break-word;
        word-break: break-word;
    }
</style>