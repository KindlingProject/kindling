//
// Created by Daxin Wang on 2023/3/17.
//

#include "utils.h"
#include <cstdio>
#include <cstdlib>
#include <time.h>

char *date = (char *)(malloc(sizeof(char) * 30));
void printCurrentTime() {
  time_t now = time(nullptr);
  strftime(date, 30, "%Y/%m/%d %H:%M:%S", localtime(&now));
  printf("%s ", date);
}

uint32_t digits10_v3(uint64_t v) {
    if (v < 10) return 1;
    if (v < 100) return 2;
    if (v < 1000) return 3;
    if (v < 1000000000000) {    // 10^12
        if (v < 100000000) {    // 10^7
            if (v < 1000000) {  // 10^6
                if (v < 10000) return 4;
                return 5 + (v >= 100000); // 10^5
            }
            return 7 + (v >= 10000000); // 10^7
        }
        if (v < 10000000000) {  // 10^10
            return 9 + (v >= 1000000000); // 10^9
        }
        return 11 + (v >= 100000000000); // 10^11
    }
    return 12 + digits10_v3(v / 1000000000000); // 10^12
}

uint32_t u64ToAscii_v3(uint64_t value, char* dst) {
    static const char digits[] =
        "0001020304050607080910111213141516171819"
        "2021222324252627282930313233343536373839"
        "4041424344454647484950515253545556575859"
        "6061626364656667686970717273747576777879"
        "8081828384858687888990919293949596979899";
 
    const size_t length = digits10_v3(value);
    uint32_t next = length - 1;
 
    while (value >= 100) {
        const uint32_t i = (value % 100) * 2;
        value /= 100;
        dst[next - 1] = digits[i];
        dst[next] = digits[i + 1];
        next -= 2;
    }
    // Handle last 1-2 digits
    if (value < 10) {
        dst[next] = '0' + uint32_t(value);
    } else {
        uint32_t i = uint32_t(value) * 2;
        dst[next - 1] = digits[i];
        dst[next] = digits[i + 1];
    }
    dst[length] = '\0';
    return length;
}

