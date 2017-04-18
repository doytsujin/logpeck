#!/bin/bash

function Usage() {
 echo "Usage:"
  echo "  $0 <task.config> [add|remove|stop|start|update]"
}

if [ $# != 2 ]; then
	Usage; exit 1
fi

conf_file=$1
if [ ! -f $conf_file ]; then 
	echo file [$conf_file] not exist.
	Usage; exit 2
fi

source $conf_file
cmd=$2
case $2 in
 	add|remove|stop|start|update)
	 	;;
 	*)
	 	Usage; exit 1
	 	;;
esac

echo $url/$cmd
echo $config
curl -XPOST "$url/peck_task/$cmd" -d "$config"
