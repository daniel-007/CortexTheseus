#include <stdint.h>
#define result_t uint32_t
#ifdef __cplusplus
extern "C"
{
#endif
    uint8_t CuckooSolve(uint8_t *header, uint32_t header_len, uint32_t nonce, result_t *result, uint32_t *result_len, uint8_t *target, uint8_t *hash);
    uint8_t CuckooVerify(uint8_t *header, uint32_t header_len, uint32_t nonce, result_t *result, uint8_t* target, uint8_t* hash);
    int32_t CuckooVerifySolutions(uint8_t *header, uint32_t header_len, result_t *result);
    int32_t CuckooVerifyHeaderNonceAndSolutions(uint8_t *header, uint32_t header_len, uint32_t nonce, result_t *result);
    void CuckooInit(uint32_t nthread);
    void CuckooFinalize();
#ifdef __cplusplus
}
#endif