<style scoped>

</style>
<template>
    <Row :gutter="16">
        <Col span="4" class="left">
            <Card>
                <div>
                    <Input v-model="inputKey" placeholder="redis key..." style="width: 100%"></Input>
                </div>
                <ul>
                    <li v-for="item in keys"><span class="layout-text" ><router-link :to=" '/serverid/'+ $route.params.serverid +'/keys/'+ item">{{item}}</router-link></span></li>
                </ul>
            </Card>
        </Col>
        <Col span="20">
            <Card>
                <router-view></router-view>
            </Card>
        </Col>
    </Row>
</template>

<script>
    export default {
        data(){
            return {
                inputKey:"",
                serverid:0,
                keys: this.getKeys(),
            }
        },
        watch: {
            '$route': 'reload',
            // 如果 question 发生改变，这个函数就会运行
            inputKey: function (newKey) {
                this.getKeys();
            }
        },
        methods: {
            getKeys: _.debounce(function(){
                var info = {}
                info.serverid = parseInt(this.$route.params.serverid);
                info.data = this.inputKey;
                this.$socket.emit("ScanKeys", info)
            }, 500),
            reload:  function(newRouter, oldRouter){
                console.log("reloading....")
                console.dir(this.$route.params)
                this.serverid = this.$route.params.serverid;
                this.inputKey = "";
                this.getKeys();
            },
        },
        socket:{
            events:{
                LoadKeys(keys){
                    this.keys = keys
                }
            }
        }

       
    }
</script>