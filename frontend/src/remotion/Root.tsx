import { Composition } from 'remotion';
import { DeveloperIntegrationDemo } from './DeveloperIntegrationDemo';

/**
 * Remotion Root — registers all compositions for rendering.
 * Used by `npx remotion render` and Remotion Studio.
 */
export const RemotionRoot = () => {
  return (
    <>
      <Composition
        id="DeveloperIntegrationDemo"
        component={DeveloperIntegrationDemo}
        durationInFrames={2190}
        fps={30}
        width={1280}
        height={720}
        defaultProps={{ narration: true }}
      />
    </>
  );
};
