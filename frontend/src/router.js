const routers = [{
    path: '/',
    meta: {
        title: 'redisSky manager'
    },
    component: (resolve) => require(['./views/index.vue'], resolve),
    children: [
    	{
    		path: '/keys',
    		meta: {
		        title: 'keys - redisSky manager'
		    },
		    component: (resolve) => require(['./views/keys.vue'], resolve),
		    children: [
		    	{
		    		path: '/keys/:key',
		    		component: (resolve) => require(['./views/item.vue'], resolve)
		    	}
		    ]
    	}
    ]
}];
export default routers;