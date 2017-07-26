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
                cmdLog(data){
                    console.log("log:", data);
                },
                cmdErr(data){
                    this.$Message.error(data, 2000);
                },
                cmdReceive(data){
                    console.log("receive:", data)
                },
                ShowServers(servers){
                    this.$store.dispatch('saveServers', servers);
                }
            }
        }
    };
</script>