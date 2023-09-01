import React, { useState } from 'react';
import { List, Form, Input, Button, message } from 'antd';
import CustomDatePicker from '@/containers/components/customDatePicker';
import _ from 'lodash';
import './index.less';

import { getCauseTraceList } from '@/request/index';

function TraceLsit() {
    const [form] = Form.useForm();
    const [traceList, setTraceList] = useState([]);

    const onSearch = () => {
        form.validateFields().then(value => {
            console.log(value);
            const params = {
                start: new Date(value.time.from).getTime() * 1000000,
                end: new Date(value.time.to).getTime() * 1000000,
                // start: value.start,
                // end: value.end,
                pid: value.pid,
                url: value.pid,
                traceId: value.traceId
            }
            getCauseTraceList(params).then(res => {
                if (res.data.data.length === 0) {
                    message.info('返回的TraceId List为空');
                }
                setTraceList(res.data.data);
            })
        })
    }

    return (
        <div className='cause_tracelist_warp'> 
            <div className='top_form_warp'>
                <Form form={form} layout='inline'>
                    <Form.Item name='time' valuePropName='timeInfo' label='time' className='custom_timewarp'>
                        <CustomDatePicker showNavgite={false}/>
                    </Form.Item>
                    {/* <Form.Item name='start' label='start'>
                        <Input/>
                    </Form.Item>
                    <Form.Item name='end' label='end'>
                        <Input/>
                    </Form.Item> */}
                    <Form.Item name='pid' label='pid'>
                        <Input/>
                    </Form.Item>
                    <Form.Item name='url' label='url'>
                        <Input/>
                    </Form.Item>
                    <Form.Item name='traceId' label='traceId'>
                        <Input/>
                    </Form.Item>
                    <Form.Item>
                        <Button onClick={onSearch}>搜索</Button>
                    </Form.Item>
                </Form>
            </div>
            <div className='tracelist_warp'>
                {
                    traceList.length > 0 ? <List size='small' bordered dataSource={traceList} renderItem={(item: any) => 
                        <List.Item>
                            <a href={`#/rootCause?traceId=${item.traceId}`}>{item.traceId}</a>
                        </List.Item>}/> : null
                }
            </div>
        </div>  
    );
}

export default TraceLsit;
