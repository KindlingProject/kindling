import React from 'react';
import { Descriptions } from 'antd'; 

interface IProps {
    showTitle?: boolean;
    data: {
        [key in string]: any
    }
}
function TextTag({data, showTitle}: IProps) {
    return (
        <React.Fragment>
            {
                Object.keys(data).length > 0 ? <Descriptions title={showTitle ? '上下文信息' : ''}>
                    {
                        Object.keys(data).map((key, idx) => <Descriptions.Item label={key}>{data[key]}</Descriptions.Item>)
                    }
                </Descriptions> : null
            }
        </React.Fragment>
        
    );
}

export default TextTag;
