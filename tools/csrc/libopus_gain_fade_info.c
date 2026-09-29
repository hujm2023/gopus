/* Float gain_fade oracle: libopus 1.6.1 src/opus_encoder.c.
 * Keep the function's assignment boundaries: native compiler contraction is
 * part of this oracle, using the reference build's -O3 -DNDEBUG flags.
 */
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "config.h"
#include "celt/arch.h"
#include "celt/modes.h"

static void gain_fade(const opus_res *in, opus_res *out, opus_val16 g1, opus_val16 g2,
        int overlap48, int frame_size, int channels, const celt_coef *window, opus_int32 Fs)
{
    int i;
    int inc;
    int overlap;
    int c;
    inc = IMAX(1, 48000/Fs);
    overlap=overlap48/inc;
    if (channels==1)
    {
       for (i=0;i<overlap;i++)
       {
          opus_val16 g, w;
          w = COEF2VAL16(window[i*inc]);
          w = MULT16_16_Q15(w, w);
          g = SHR32(MAC16_16(MULT16_16(w,g2),
                Q15ONE-w, g1), 15);
          out[i] = MULT16_RES_Q15(g, in[i]);
       }
    } else {
       for (i=0;i<overlap;i++)
       {
          opus_val16 g, w;
          w = COEF2VAL16(window[i*inc]);
          w = MULT16_16_Q15(w, w);
          g = SHR32(MAC16_16(MULT16_16(w,g2),
                Q15ONE-w, g1), 15);
          out[i*2] = MULT16_RES_Q15(g, in[i*2]);
          out[i*2+1] = MULT16_RES_Q15(g, in[i*2+1]);
       }
    }
    c=0;do {
       for (i=overlap;i<frame_size;i++)
       {
          out[i*channels+c] = MULT16_RES_Q15(g2, in[i*channels+c]);
       }
    }
    while (++c<channels);
}

int main(int argc, char **argv) {
    int rate, channels, frame_size, err = 0;
    float previous, current;
    opus_res samples[1920];
    CELTMode *mode;
    if (argc != 5) return 1;
    rate = atoi(argv[1]);
    channels = atoi(argv[2]);
    previous = strtof(argv[3], NULL);
    current = strtof(argv[4], NULL);
    if ((rate != 48000 && rate != 24000) ||
        (channels != 1 && channels != 2)) return 1;
    frame_size = rate / 50;
    mode = opus_custom_mode_create(48000, 960, &err);
    if (mode == NULL) return 1;
    for (int i = 0; i < frame_size * channels; i++) samples[i] = .25f;
    gain_fade(samples, samples, previous, current, 120, frame_size,
              channels, mode->window, rate);
    for (int i = 0; i < frame_size * channels; i++) {
        uint32_t bits;
        memcpy(&bits, &samples[i], sizeof(bits));
        printf("%08x\n", (unsigned int)bits);
    }
    return ferror(stdout) ? 1 : 0;
}
