import type ytdl from 'ytdl-core';

const desiredQualityOrder = [
  'AUDIO_QUALITY_MEDIUM',
  'AUDIO_QUALITY_LOW',
  'AUDIO_QUALITY_HIGH',
] as const;

export function findDesiredFormat(videoInfo: ytdl.videoInfo) {
  const audioFormats = videoInfo.formats.filter(format =>
    format.mimeType?.startsWith('audio')
  );

  let bestFormats: ytdl.videoFormat[] = [];

  for (const desiredQuality of desiredQualityOrder) {
    const formatsForQuality = audioFormats.filter(
      format => format.audioQuality === desiredQuality
    );

    if (formatsForQuality.length) {
      bestFormats = formatsForQuality;

      break;
    }
  }

  return bestFormats.reduce((acc, format) => {
    const formatSize = parseFloat(format.contentLength);
    const accSize = parseFloat(acc.contentLength);

    return formatSize < accSize ? format : acc;
  }, audioFormats[0]);
}
