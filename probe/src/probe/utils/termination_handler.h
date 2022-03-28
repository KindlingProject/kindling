//
// Created by 千陆 on 2022/1/6.
//

#ifndef KINDLING_PROBE_KINDLING_MANAGER_H
#define KINDLING_PROBE_KINDLING_MANAGER_H

#include <src/stirling/stirling.h>
#include <src/common/base/base.h>
#include <src/common/signal/signal.h>
#include "sinsp.h"
#include <memory>

class TerminationHandler {
public:
    static constexpr auto kSignals = px::MakeArray(SIGINT, SIGQUIT, SIGTERM, SIGHUP);

    static void InstallSignalHandlers() {
        for (size_t i = 0; i < kSignals.size(); ++i) {
            signal(kSignals[i], TerminationHandler::OnTerminate);
        }
    }

    static void set_stirling(px::stirling::Stirling* stirling_) { m_stirling_ = stirling_; }
    static void set_sinsp(sinsp* sinsp_) { m_sinsp_ = sinsp_; }

    static void OnTerminate(int signum) {
        if (m_sinsp_ != nullptr) {
            LOG(INFO) << "Trying to gracefully stop sinsp";
            m_sinsp_->close();
        }
        if (m_stirling_ != nullptr) {
            LOG(INFO) << "Trying to gracefully stop stirling";
            m_stirling_->Stop();
        }
        exit(signum);
    }

private:
    inline static px::stirling::Stirling* m_stirling_ = nullptr;
    inline static sinsp* m_sinsp_ = nullptr;
};



#endif //KINDLING_PROBE_KINDLING_MANAGER_H
