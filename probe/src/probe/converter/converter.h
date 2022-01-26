//
// Created by 散养鸡 on 2021/12/13.
//

#ifndef HCMINE_CONVERT_H
#define HCMINE_CONVERT_H
#include <string>
//#include <publisher/publisher.h>
#include "src/probe/converter/kindling_event.pb.h"

class converter {
public:
	// source evt -> kindling evt
	virtual void convert(kindling::KindlingEvent* kevt, void * evt) = 0;
    std::string GetName();
private:
    std::string m_name;
//    publisher *m_publisher;
};

#endif //HCMINE_CONVERT_H
