type SeriesSize = 'sm' | 'md' | 'lg';

export interface SimpleOptions {
  layout: string;
  showLatency: boolean;
  normalLatency: number;
  abnormalLatency: number;
  showRtt: boolean;
  normalRtt: number;
  abnormalRtt: number;
  seriesCountSize: SeriesSize;
}
