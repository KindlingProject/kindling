if [ ! -n "$(ps -ef|grep kindling-collector |awk '$0 !~/grep/ {print $1}' |tr -s '\n' ' ')" ]; then
	echo "no such process"
else
	kill -9 $(ps -ef|grep kindling-collector |awk '$0 !~/grep/ {print $1}' |tr -s '\n' ' ')