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
                    <li v-for="item in keys"><span class="layout-text" ><router-link :to="'/keys/'+ item">{{item}}</router-link></span></li>
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
                keys:[]
            }
        },
        watch: {
            // 如果 question 发生改变，这个函数就会运行
            inputKey: function (newKey) {
                this.getKeys();
            }
        },
        methods: {
            getKeys: _.debounce(function(){
                if (this.inputKey == "") {
                    this.keys = [];
                }else{
                    this.keys = ["test1","test2"];
                }
            }, 500)
        }

       
    }
</script>