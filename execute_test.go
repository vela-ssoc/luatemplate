package luatemplate

import (
	"os"
	"testing"
)

func TestExecute(t *testing.T) {
	tpl, err := New().Parse(exampleTpl)
	if err != nil {
		t.Fatal(err)
	}

	arg := tpl.ParamJSON()
	t.Log(arg)

	param := map[string]any{
		"user":           "root",
		"process_number": "4",
		"error": map[string]string{
			"path": "/tmp/error.log",
		},
	}
	err = tpl.Execute(os.Stdout, param)
	t.Log(err)
}

const exampleTpl = `
local t = gee.template{
    version = "0.0.1",
    auth    = "vela",
}

local p = gee.param

local init_by_lua_block = [[

local errlog = require "ngx.errlog"
local status, err = errlog.set_filter_level(ngx.ERR)
if not status then
    ngx.log(ngx.ERR, err)
    return
end

]]

local init_worker_by_lua_block = [[

local worker = require("rock.worker")
worker.sync()
]]

local upstream_backend_lua = [[

server 10.205.121.75:80;
balancer_by_lua_file rock.d/balancer;
keepalive 36;
]]

t.param = {
    p.string("user").must("root" , "www" , "rock").label("进程用户").default("root"),

    p.map("error").label("错误日志").table{
        p.string("path").default("logs/error_log").label("日志路径").style{span=16, label="inner"},
        p.string("level").label("日志等级").default("error").must("debug" , "error" , "notice").style{span=8, label="inner"},
    },

    p.string("process_number").default("auto").label("进程数"),
    p.string("pid").default("logs/rock.pid").label("进程PID"),
    p.string("worker_rlimit_nofile").default("65535").must(1024,4096,65535).label("进程句柄"),

    p.map("events").label("进程参数").size(1).table{
        p.string("worker_connections").label("进程连接数:").default(65535).must(1024,4096,65535).style{span=24},
    },

    p.array("shm").label("共享内存").table{
        p.string("name").label("名称").style{span=10, label="inner"},
        p.string("value").label("大小").style{span=14, label="inner"},
    }.default{
        {name="upstream" ,value="2m"},
        {name="index"    ,value="2m"},
        {name="logger"   ,value="100m"},
        {name="status"   ,value="50m"},
        {name="pool"     ,value="100m"},
        {name="black"    ,value="200m"},
        {name="limit"    ,value="1000m"},
        {name="rtx"      ,value="10m"},
    },

    p.array("http").label("web服务").table{
        p.string("name").label("字段名").style{span=10 , label="inner"},
        p.string("value").label("字段值").style{span=14 , label="inner"},
    }.default{
        {name="log_format" , value="main $json"},
        {name="access_log" , value="/vdb/logs/access.$year-$month-$day.$hour.log.ts  main" },
        {name="proxy_ignore_client_abort" , value="on"} ,
        {name="lua_socket_log_errors" , value="off" },
        {name="underscores_in_headers" , value="on"},
        {name="server_tokens" , value="off"},
        {name="sendfile" , value="on"},
        {name="tcp_nopush" , value="on"},
        {name="tcp_nodelay" , value="on"},
        {name="sendfile_max_chunk" , value="512k"},
        {name="keepalive_timeout" , value="120"},
        {name="keepalive_requests" , value="20000"},
        {name="default_type" , value="application/octet-stream"},
        {name="client_max_body_size" , value="3m"},
        {name="gzip" , value="on"},
        {name="gzip_types" , value="*"},
        {name="gzip_http_version" , value="1.0"},
        {name="gzip_disable" , value="'MSIE [1-6]\\.'"},
        {name="gzip_min_length" , value="5"},
        {name="gzip_comp_level" , value="9"},
        {name="gzip_buffers" , value="4 16k"},
        {name="gzip_vary" , value="on"},
        {name="lua_package_path" , value="/usr/local/rock/?.lua;;"},
        {name="lua_capture_error_log" , value="20m"},
        {name="lua_code_cache" , value="on"},
        {name="lua_max_running_timers" , value="1024"},
        {name="lua_max_pending_timers" , value="1024"},
    },

    p.lua("init").label("启动设置").desc("init_by_lua_block").default(init_by_lua_block).style{span=16},
    p.lua("init_worker").label("工作进程").desc("init_worker_by_lua_block").default(init_worker_by_lua_block).style{span=16},
    p.lua("backend").label("默认上游").desc("upstream backend").default(upstream_backend_lua).style{span=16}

}

t.text = [[
user  {{.user}};
worker_processes  {{.process_number}};
error_log  {{.error.path}} {{.error.level}};
pid        {{.pid}};
worker_rlimit_nofile {{.worker_rlimit_nofile}};

events {
    worker_connections  {{.events.worker_connections}};
}

http {

    include mime.types;
    {{array .http "%s %s;\n" "name" "value" | indent 4 }}

    {{array .shm  "lua_shared_dict %s %s;\n" "name" "value" | indent 4 }}

    # init
    init_by_lua_block {
    {{lua .init | indent 8 }}
    }

    # update
    init_worker_by_lua_block {
    {{lua .init_worker | indent 8 }}
    }

    # upstream backend
    upstream backend {
    {{lua .backend | indent 8 }}
    }


    # rewrite
    rewrite_by_lua_file rock.d/rewrite;

    # security access
    access_by_lua_file rock.d/security;

    # security header_filter
    header_filter_by_lua_file rock.d/security;

    # security body_filter
    body_filter_by_lua_file rock.d/security;

    # security log
    log_by_lua_file rock.d/security;



    # servers app cfg
    include server.d/*.conf;
}

]]
`
