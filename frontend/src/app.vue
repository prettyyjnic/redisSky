<template>
    <div>
        <router-view></router-view>
    </div>
</template>
<script>
    export default {
        data() {
            return {
                loading: null
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
                loading () {
                    this.loading = this.$Message.loading({
                        content: 'loading ...',
                        duration: 0
                    });
                },
                loadingComplete(){
                    this.loading();
                },
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