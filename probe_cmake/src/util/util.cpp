#include "utils.h"

void kindling_strcpy(char *d, char *s, int len) {
	int i = 0;
	while (i < len) {
		i++;
		*d++ = *s++;
	}
	*d = '\0';
}