#ifndef LIBMIAOSIC_H
#define LIBMIAOSIC_H

#ifdef __cplusplus
extern "C" {
#endif

#include <stdint.h>

typedef struct {
	char* provider;
	char* identifier;
} MiaosicMetaData;

typedef struct {
	char* url;
	unsigned char* data;
	int data_len;
} MiaosicPicture;

typedef struct {
	char* title;
	char* artist;
	MiaosicPicture cover;
	char* album;
	MiaosicMetaData meta;
} MiaosicMediaInfo;

typedef struct {
	char* key;
	char* value;
} MiaosicHeaderPair;

typedef struct {
	char* url;
	char* quality;
	MiaosicHeaderPair* headers;
	int header_len;
} MiaosicMediaUrl;

typedef struct {
	int len;
	MiaosicMediaInfo* items;
} MiaosicMediaInfoList;

typedef struct {
	int len;
	MiaosicMediaUrl* items;
} MiaosicMediaUrlList;

typedef struct {
	char* title;
	MiaosicMediaInfoList medias;
	MiaosicMetaData meta;
} MiaosicPlaylist;

typedef struct {
	int matched;
	MiaosicMetaData meta;
} MiaosicMatchResult;

typedef struct {
	char* url;
	char* key;
} MiaosicQrLoginSession;

typedef struct {
	int success;
	char* message;
} MiaosicQrLoginResult;

typedef struct {
	char* lang;
	char* lyrics;
} MiaosicLyrics;

typedef struct {
	int len;
	MiaosicLyrics* items;
} MiaosicLyricsList;

typedef struct {
	int value;
} MiaosicBool;

typedef struct {
	char* value;
} MiaosicString;

typedef struct {
	int len;
	char** items;
} MiaosicStringList;

typedef enum {
	MIAOSIC_RESULT_NONE = 0,
	MIAOSIC_RESULT_BOOL = 1,
	MIAOSIC_RESULT_STRING = 2,
	MIAOSIC_RESULT_STRING_LIST = 3,
	MIAOSIC_RESULT_META = 4,
	MIAOSIC_RESULT_MEDIA_INFO = 5,
	MIAOSIC_RESULT_MEDIA_INFO_LIST = 6,
	MIAOSIC_RESULT_MEDIA_URL_LIST = 7,
	MIAOSIC_RESULT_PLAYLIST = 8,
	MIAOSIC_RESULT_MATCH = 9,
	MIAOSIC_RESULT_QR_LOGIN_SESSION = 10,
	MIAOSIC_RESULT_QR_LOGIN_RESULT = 11,
	MIAOSIC_RESULT_LYRICS_LIST = 12
} MiaosicResultType;

typedef struct {
	int ok;
	char* err;
	MiaosicResultType result_type;
	void* data;
} MiaosicResult;

void FreeResult(MiaosicResult* res);

MiaosicResult* SearchByProvider(const char* provider, const char* keyword, int page, int size);
MiaosicResult* GetMediaUrl(const char* provider, const char* identifier, const char* quality);
MiaosicResult* GetMediaInfo(const char* provider, const char* identifier);
MiaosicResult* GetMediaLyric(const char* provider, const char* identifier);
MiaosicResult* MatchPlaylistByProvider(const char* provider, const char* uri);
MiaosicResult* GetPlaylist(const char* provider, const char* identifier);
MiaosicResult* MatchMedia(const char* keyword);
MiaosicResult* MatchMediaByProvider(const char* provider, const char* uri);
MiaosicResult* ListAvailableProviders(void);

MiaosicResult* LoginByProvider(const char* provider, const char* username, const char* password);
MiaosicResult* LogoutByProvider(const char* provider);
MiaosicResult* IsLoginByProvider(const char* provider);
MiaosicResult* RefreshLoginByProvider(const char* provider);
MiaosicResult* QrLoginByProvider(const char* provider);
MiaosicResult* QrLoginVerifyByProvider(const char* provider, const char* key, const char* url);
MiaosicResult* RestoreSessionByProvider(const char* provider, const char* session);
MiaosicResult* SaveSessionByProvider(const char* provider);

MiaosicResult* UseBilibiliVideo(void);
MiaosicResult* UseKugou(void);
MiaosicResult* UseKugouInstrumental(void);
MiaosicResult* UseKuwo(void);
MiaosicResult* UseNetease(void);
MiaosicResult* UseQQLogin(void);
MiaosicResult* UseWechatLogin(void);
MiaosicResult* UseLocal(const char* local_dir);

#ifdef __cplusplus
}
#endif

#endif
