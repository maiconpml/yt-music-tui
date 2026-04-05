package goytmusic

import "strings"

const (
	// json paths fragments to extract data from API responses
	pSingleColumn                = "contents.singleColumnBrowseResultsRenderer"
	pTwoColumn                   = "contents.twoColumnBrowseResultsRenderer"
	pGridRendererItems           = "gridRenderer.items"
	pTab0                        = "tabs.0"
	pContent0                    = "contents.0"
	pContents                    = "contents"
	pTabRendererContent          = "tabRenderer.content"
	pSectionList                 = "sectionListRenderer"
	pMusicResponsiveHeader       = "musicResponsiveHeaderRenderer"
	pMusicEditablePlaylistHeader = "musicEditablePlaylistDetailHeaderRenderer.header"
	pSecContents                 = "secondaryContents"
	pPlaylistShelf               = "musicPlaylistShelfRenderer"
	pMusicShelf                  = "musicShelfRenderer"
	pRespListItem                = "musicResponsiveListItemRenderer"
	pRespListItemFlexColumn      = "musicResponsiveListItemFlexColumnRenderer"
	pFlexColumn0                 = "flexColumns.0"
	pFlexColumn1                 = "flexColumns.1"
	pFlexColumn2                 = "flexColumns.2"
	pMusicTwoRow                 = "musicTwoRowItemRenderer"
	pRun0                        = "runs.0"
	pRun2                        = "runs.2"
	pRun4                        = "runs.4"
	pRuns                        = "runs"
	pText                        = "text"
	pNavEndpoint                 = "navigationEndpoint"
	pBrowseEnd                   = "browseEndpoint"
	pBrowseID                    = "browseId"
	pWatchEnd                    = "watchEndpoint"
	pVideoID                     = "videoId"
	pPlaylistID                  = "playlistId"
	pPlaylistSetVideoID          = "playlistSetVideoId"
	pTitle                       = "title"
	pSubtitle                    = "subtitle"
	pFacepileStackView           = "facepile.avatarStackViewModel"
	pRendCtxtInnertubeCommand    = "rendererContext.commandContext.onTap.innertubeCommand"
	pContent                     = "content"
	pOverlayRenderer             = "overlay.musicItemThumbnailOverlayRenderer"
	pMusicPlayButtonRenderer     = "musicPlayButtonRenderer"
	pPlayNavEndpoint             = "playNavigationEndpoint"

	pSingleColumnNextRts               = "contents.singleColumnMusicWatchNextResultsRenderer"
	pTabbedRenderer                    = "tabbedRenderer.watchNextTabbedResultsRenderer"
	pMusicQueueRenderer                = "musicQueueRenderer"
	pPlaylistPanelRenderer             = "playlistPanelRenderer"
	pPlaylistPanelVideoWrapperRenderer = "playlistPanelVideoWrapperRenderer"
	pPrimaryRenderer                   = "primaryRenderer"
	pPlaylistPanelVideoRenderer        = "playlistPanelVideoRenderer"
	pLongByLineText                    = "longBylineText"
	pBrowseEndContextPageType          = "browseEndpointContextSupportedConfigs.browseEndpointContextMusicConfig.pageType"
	pContinuation                      = "continuations.0.nextContinuationData.continuation"
	pLengthText                        = "lengthText"
)

// joinPaths takes multiple strings s1, s2, ..., sn and join
// them in s1.s2....sn
func joinPaths(parts ...string) string {
	return strings.Join(parts, ".")
}
