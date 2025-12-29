#include <stdio.h>
#include <stdlib.h>

#include "libmiaosic.h"

static void print_error(const char* label, MiaosicResult* res) {
    if (res == NULL) {
        fprintf(stderr, "%s: result is NULL\n", label);
        return;
    }
    if (res->ok) {
        return;
    }
    fprintf(stderr, "%s failed: %s\n", label, res->err ? res->err : "unknown error");
}

int main(void) {
    MiaosicResult* reg = UseBilibiliVideo();
    if (reg == NULL || !reg->ok) {
        print_error("UseBilibiliVideo", reg);
        FreeResult(reg);
        return 1;
    }
    FreeResult(reg);

    MiaosicResult* search = SearchByProvider("bilibili-video", "家有女友", 1, 5);
    if (search == NULL || !search->ok || search->result_type != MIAOSIC_RESULT_MEDIA_INFO_LIST) {
        print_error("SearchByProvider", search);
        FreeResult(search);
        return 1;
    }

    MiaosicMediaInfoList* list = (MiaosicMediaInfoList*)search->data;
    printf("Search results: %d\n", list ? list->len : 0);
    if (list && list->len > 0) {
        for (int i = 0; i < list->len; i++) {
            MiaosicMediaInfo* info = &list->items[i];
            const char* title = info->title ? info->title : "";
            const char* artist = info->artist ? info->artist : "";
            const char* provider = info->meta.provider ? info->meta.provider : "";
            const char* identifier = info->meta.identifier ? info->meta.identifier : "";
            printf("[%d] %s - %s (%s:%s)\n", i + 1, title, artist, provider, identifier);
        }
    }

    FreeResult(search);
    return 0;
}
