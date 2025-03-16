# A RC script for freebsd frpc

Today I'm trying to setup a jumpserver for my local network using frpc. My jumpserver is a Linux machine before but
now I decide to use FreeBSD instead.

Edit the rc script file `/usr/local/etc/rc.d/frpc` first:

```bash
#!/bin/sh
#
# PROVIDE: frpc
# REQUIRE: NETWORKING SERVERS
# KEYWORD: shutdown
#
# Add the following lines to /etc/rc.conf.local or /etc/rc.conf
# to enable this service:
#
# frpc_enable (bool):        Set to YES to enable frpc
#                            Default: NO
# frpc_config (str):         Main configuration file
#                            Default: /usr/local/etc/frpc/frpc.ini
# frpc_user (str):           User to run frpc as
#                            Default: nobody
# frpc_group (str):          Group to run frpc as
#                            Default: nobody
# frpc_instances (str):      List of named instances of frpc
#                            Default: empty
# 
# For each named instance in frpc_instances, the following variables
# can be defined to override the defaults:
#
# frpc_${instance}_enable (bool):    Set to YES to enable this instance
#                                    Default: NO
# frpc_${instance}_config (str):     Configuration file for this instance
#                                    Default: /usr/local/etc/frpc/frpc_${instance}.ini
# frpc_${instance}_user (str):       User to run this instance as
#                                    Default: nobody
# frpc_${instance}_group (str):      Group to run this instance as
#                                    Default: nobody
# frpc_${instance}_flags (str):      Additional flags for this instance
#                                    Default: empty
#
# Example:
#   frpc_enable="YES"
#   frpc_instances="office home"
#   frpc_office_enable="YES"
#   frpc_office_config="/usr/local/etc/frpc/frpc_office.ini"
#   frpc_home_enable="YES"
#   frpc_home_config="/usr/local/etc/frpc/frpc_home.ini"
#

. /etc/rc.subr

name="frpc"
rcvar=frpc_enable

load_rc_config $name

: ${frpc_enable:="NO"}
: ${frpc_config:="/usr/local/etc/frpc/frpc.ini"}
: ${frpc_user:="nobody"}
: ${frpc_group:="nobody"}
: ${frpc_flags:=""}
: ${frpc_instances:=""}

command="/usr/local/bin/frpc"
piddir="/var/run/frpc"
pidfile="${piddir}/${name}.pid"

# Create pid directory if it doesn't exist
if [ ! -d ${piddir} ]; then
    mkdir -p ${piddir}
    if [ -n "${frpc_user}" ]; then
        chown ${frpc_user}:${frpc_group} ${piddir}
    fi
    chmod 755 ${piddir}
fi

start_cmd="frpc_start"
stop_cmd="frpc_stop"
status_cmd="frpc_status"
restart_cmd="frpc_restart"
configtest_cmd="frpc_configtest"
extra_commands="status restart configtest"

frpc_start()
{
    if [ -n "$1" ]; then
        _instance_start "$1"
    else
        # Start the main instance if enabled
        if checkyesno frpc_enable; then
            echo "Starting frpc."
            _instance_start ""
        fi
        
        # Start all enabled instances
        for instance in ${frpc_instances}; do
            eval frpc_instance_enable=\$frpc_${instance}_enable
            if checkyesno frpc_instance_enable; then
                echo "Starting frpc instance: ${instance}"
                _instance_start "${instance}"
            fi
        done
    fi
}

frpc_stop()
{
    if [ -n "$1" ]; then
        _instance_stop "$1"
    else
        # Stop all enabled instances
        for instance in ${frpc_instances}; do
            eval frpc_instance_enable=\$frpc_${instance}_enable
            if checkyesno frpc_instance_enable; then
                echo "Stopping frpc instance: ${instance}"
                _instance_stop "${instance}"
            fi
        done
        
        # Stop the main instance if enabled
        if checkyesno frpc_enable; then
            echo "Stopping frpc."
            _instance_stop ""
        fi
    fi
}

frpc_status()
{
    if [ -n "$1" ]; then
        _instance_status "$1"
    else
        # Check status of main instance if enabled
        if checkyesno frpc_enable; then
            _instance_status ""
        fi
        
        # Check status of all enabled instances
        for instance in ${frpc_instances}; do
            eval frpc_instance_enable=\$frpc_${instance}_enable
            if checkyesno frpc_instance_enable; then
                _instance_status "${instance}"
            fi
        done
    fi
}

frpc_restart()
{
    if [ -n "$1" ]; then
        _instance_restart "$1"
    else
        frpc_stop
        frpc_start
    fi
}

frpc_configtest()
{
    if [ -n "$1" ]; then
        _instance_configtest "$1"
    else
        # Test main instance config if enabled
        if checkyesno frpc_enable; then
            _instance_configtest ""
        fi
        
        # Test all enabled instances
        for instance in ${frpc_instances}; do
            eval frpc_instance_enable=\$frpc_${instance}_enable
            if checkyesno frpc_instance_enable; then
                _instance_configtest "${instance}"
            fi
        done
    fi
}

_instance_start()
{
    local instance="$1"
    local inst_name
    
    if [ -n "${instance}" ]; then
        inst_name="${name}_${instance}"
        eval frpc_instance_config=\$frpc_${instance}_config
        eval frpc_instance_user=\$frpc_${instance}_user
        eval frpc_instance_group=\$frpc_${instance}_group
        eval frpc_instance_flags=\$frpc_${instance}_flags
        
        : ${frpc_instance_config:="/usr/local/etc/frpc/frpc_${instance}.ini"}
        : ${frpc_instance_user:="${frpc_user}"}
        : ${frpc_instance_group:="${frpc_group}"}
        : ${frpc_instance_flags:="${frpc_flags}"}
        
        instance_pidfile="${piddir}/${inst_name}.pid"
        instance_config="${frpc_instance_config}"
        instance_user="${frpc_instance_user}"
        instance_group="${frpc_instance_group}"
        instance_flags="${frpc_instance_flags}"
    else
        inst_name="${name}"
        instance_pidfile="${pidfile}"
        instance_config="${frpc_config}"
        instance_user="${frpc_user}"
        instance_group="${frpc_group}"
        instance_flags="${frpc_flags}"
    fi
    
    if [ ! -f "${instance_config}" ]; then
        echo "Error: ${instance_config} does not exist."
        return 1
    fi
    
    echo "Starting ${inst_name}."
    /usr/sbin/daemon -f -p ${instance_pidfile} -u ${instance_user} \
        ${command} -c ${instance_config} ${instance_flags}
}

_instance_stop()
{
    local instance="$1"
    local inst_name
    
    if [ -n "${instance}" ]; then
        inst_name="${name}_${instance}"
        instance_pidfile="${piddir}/${inst_name}.pid"
    else
        inst_name="${name}"
        instance_pidfile="${pidfile}"
    fi
    
    if [ -f "${instance_pidfile}" ]; then
        echo "Stopping ${inst_name}."
        kill `cat ${instance_pidfile}` 2>/dev/null
        rm -f ${instance_pidfile}
    else
        echo "${inst_name} is not running."
    fi
}

_instance_status()
{
    local instance="$1"
    local inst_name
    
    if [ -n "${instance}" ]; then
        inst_name="${name}_${instance}"
        instance_pidfile="${piddir}/${inst_name}.pid"
    else
        inst_name="${name}"
        instance_pidfile="${pidfile}"
    fi
    
    if [ -f "${instance_pidfile}" ]; then
        pid=`cat ${instance_pidfile}`
        if ps -p ${pid} > /dev/null 2>&1; then
            echo "${inst_name} is running as pid ${pid}."
            return 0
        else
            echo "${inst_name} is not running but pidfile exists."
            return 1
        fi
    else
        echo "${inst_name} is not running."
        return 1
    fi
}

_instance_restart()
{
    local instance="$1"
    
    _instance_stop "${instance}"
    sleep 1
    _instance_start "${instance}"
}

_instance_configtest()
{
    local instance="$1"
    local inst_name
    
    if [ -n "${instance}" ]; then
        inst_name="${name}_${instance}"
        eval frpc_instance_config=\$frpc_${instance}_config
        : ${frpc_instance_config:="/usr/local/etc/frpc/frpc_${instance}.ini"}
        instance_config="${frpc_instance_config}"
    else
        inst_name="${name}"
        instance_config="${frpc_config}"
    fi
    
    if [ ! -f "${instance_config}" ]; then
        echo "Error: ${instance_config} does not exist."
        return 1
    fi
    
    echo "Testing configuration for ${inst_name}:"
    ${command} verify -c ${instance_config}
}

# Handle instance-specific commands
if [ $# -gt 0 ]; then
    # Extract instance name if command contains colon
    case "$1" in
        *:*)
            instance="${1#*:}"
            command="${1%%:*}"
            shift
            case "${command}" in
                start|stop|restart|status|configtest)
                    frpc_${command} "${instance}" $@
                    exit $?
                    ;;
                *)
                    echo "Unknown command: ${command}"
                    exit 1
                    ;;
            esac
            ;;
    esac
fi

run_rc_command "$1"
```

Give permission by executing `chmod +x /usr/local/etc/rc.d/frpc`.

Enable frpc in `/etc/rc.conf`:

```ini
# Enable the main frpc instance
frpc_enable="YES"
frpc_config="/usr/local/etc/frpc/frpc.ini"

# Define additional instances
frpc_instances="office home"

# Configure the "office" instance
frpc_office_enable="YES"
frpc_office_config="/usr/local/etc/frpc/frpc_office.ini"

# Configure the "home" instance
frpc_home_enable="YES"
frpc_home_config="/usr/local/etc/frpc/frpc_home.ini"
```

And now we can using frpc by commands:

```bash
# Start all enabled instances:
service frpc start
# Start a specific instance:
service frpc start:office
# Stop a specific instance:
service frpc stop:home
# Check status of all instances:
service frpc status
# Check status of a specific instance:
service frpc status:office
# Test configuration of a specific instance:
service frpc configtest:home
```
