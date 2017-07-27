<template>
    <div>
        <router-view></router-view>
    </div>
</template>
<script>
    export default {
        data() {
            return {

            };
        },
        mounted() {

        },
        beforeDestroy() {

        },
        methods: {
            
        },
        socket:{
            events:{
                
                tip(type, msg){
                    this.$Message[type]({
                        content: msg,
                        duration: 10,
                        closable: true
                    })
                },
                ShowServers(servers){
                    this.$store.dispatch('saveServers', servers);
                }
            }
        }
    };
</script>