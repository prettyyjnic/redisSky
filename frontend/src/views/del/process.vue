<style scoped>
.padding3{
    padding:3px 0;
}
</style>
<template>
    <Row :gutter="16" style="min-height:800px">        
        <Col v-if="!isTaskRunning">
            <h4>no task is running...</h4>
        </Col>
        <Col v-else span="11" v-for="item in tasks" style="margin:5px;">
            <div>
                <Tooltip placement="bottom-start">                    
                    任务：{{item.id}}
                    <div slot="content">
                        <p v-for="key in item.keys"><i>{{key}}</i></p>
                    </div>
                </Tooltip>
                <Progress :status="item.status" :percent="item.process"><span v-if="item.errMsg">{{item.errMsg}}</span></Progress>            
                <Button @click="deleteTask(item.id)" type="warning" size="small">Delete</Button>
            </div>
        </Col>
    </Row>
</template>

<script>
    export default {
        data(){
            return {            
                timer: null,    
                tasks: [],
            }
        },
        computed: {
            isTaskRunning(){
                return this.tasks.length > 0 ? true : false;
            }
        },
        beforeRouteLeave(to, from, next){
            clearInterval(this.timer);
            next();
        },
        created () {
            // 组件创建完后获取数据，
            // 此时 data 已经被 observed 了
            var that = this;
            this.timer = setInterval(function(){
                that.$socket.emit("GetDelTasksProcess", []);
            }, 1000)
        },
        watch: {
            // 如果路由有变化，会再次执行该方法
        },
        methods: {
            deleteTask(taskId){
                this.$Modal.confirm({
                    content: '<p>Sure delete this task ? This could not be recoverd</p>',
                    onOk: () => {
                        this.$socket.emit("DelDeleteTask", taskId)
                    }
                });
            }
        }
        ,socket:{
            events:{
                ShowDelTaskProcess(tasks){
                    if (tasks !=null) {
                        for (var i = tasks.length - 1; i >= 0; i--) {
                            tasks[i].process = tasks[i].process * 100;
                            if (tasks[i].errMsg != '') {
                                tasks[i].status = "wrong";
                            }else if(tasks[i].process < 100){
                                tasks[i].status = "active";
                            }
                        }
                        this.tasks = tasks;
                    }
                }
            }
        }

    }
</script>