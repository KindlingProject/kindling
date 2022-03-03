## Grafana 7.X
### direct use


1. cd kindling/grafana-plugin
2. cp -r ./topo-plugin /var/lib/grafana/plugins
3. restart grafana
4. Upload JSON file  (dashborad json path => cd kindling/grafana-plugin/dashboard-json) ![image.png](https://cdn.nlark.com/yuque/0/2022/png/271213/1642734036300-fe20b3be-3701-4e0c-ad65-672638e361ab.png#clientId=u06dee3dd-ddbb-4&crop=0&crop=0&crop=1&crop=1&from=paste&height=514&id=ud75d3fd5&margin=%5Bobject%20Object%5D&name=image.png&originHeight=514&originWidth=450&originalType=binary&ratio=1&rotation=0&showTitle=false&size=51127&status=done&style=none&taskId=uaf836a04-3e58-4278-b26f-77a735b2429&title=&width=450)

## Grafana 8.X
if you not sure the path of grafana plugin, run    find ./ -name 'grafana'    to find path

1. cd kindling/grafana-plugin
2. cp -r ./topo-plugin /var/lib/grafana/plugins
3. cd /etc/grafana
4. vi grafana.init 

modify:  allow_loading_unsigned_plugins=thousand-topo-plugin
> you need delete ‘;’ before allow_loading_unsigned_plugins, otherwise it will not work.

Enter a comma-separated list of plugin identifiers to identify plugins to load even if they are unsigned. Plugins with modified signatures are never loaded.![image.png](https://cdn.nlark.com/yuque/0/2022/png/271213/1642734375690-5e14781c-0000-4340-b243-301f1928f7b6.png#clientId=u06dee3dd-ddbb-4&crop=0&crop=0&crop=1&crop=1&from=paste&height=398&id=u6af5ba8d&margin=%5Bobject%20Object%5D&name=image.png&originHeight=398&originWidth=2296&originalType=binary&ratio=1&rotation=0&showTitle=false&size=248053&status=done&style=none&taskId=u46a4f255-5f5e-462b-ad42-287b1e19d62&title=&width=2296)

5. restart grafana
6. Upload JSON file  (dashborad json path => cd kindling/grafana-plugin/dashboard-json) 
