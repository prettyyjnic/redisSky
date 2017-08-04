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
                
                tip(info){
                    this.$Message[info.type]({
                        content: info.msg,
                        duration: info.duration,
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