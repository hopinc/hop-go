package hop

import (
	"context"
	"golang.org/x/sync/errgroup"
	"net/url"

	"github.com/hopinc/hop-go/types"
)

// Create is used to create a channel. The channelType argument should be the type of channel that you want to create, state
// should be a map that you want to associate with the channel, id should be the ID you wish to specify (oe a blank string
// if you wish for this to be auto-generated), and projectId should be either a project ID to assign this to or a blank string.
func (c ClientCategoryChannels) Create(
	ctx context.Context, channelType types.ChannelType, state map[string]any, id, projectId string,
) (*types.Channel, error) {
	// Setup everything we need to do the request.
	method := "POST"
	path := "/channels"
	query := map[string]string{}
	if id != "" {
		method = "PUT"
		path += "/" + url.PathEscape(id)
		query["channel_id"] = id
	}
	if state == nil {
		state = map[string]any{}
	}
	if projectId != "" {
		query["project"] = projectId
	}

	// Do the request.
	var ch types.Channel
	err := c.c.do(ctx, clientArgs{
		method:    method,
		path:      path,
		resultKey: "channel",
		query:     query,
		body:      map[string]any{"type": channelType, "state": state},
		result:    &ch,
		ignore404: false,
	})
	if err != nil {
		return nil, err
	}
	return &ch, nil
}

func getProjectIdParam(projectId string) map[string]string {
	if projectId != "" {
		return map[string]string{"project": projectId}
	}
	return nil
}

// Get is used to get a channel. Will throw types.NotFound if it was not found.
func (c ClientCategoryChannels) Get(ctx context.Context, projectId, id string) (*types.Channel, error) {
	var ch types.Channel
	err := c.c.do(ctx, clientArgs{
		method:    "GET",
		path:      "/channels/" + url.PathEscape(id),
		resultKey: "channel",
		result:    &ch,
		query:     getProjectIdParam(projectId),
		ignore404: false,
	})
	if err != nil {
		return nil, err
	}
	return &ch, nil
}

// GetAll returns a paginator to get all the channels.
func (c ClientCategoryChannels) GetAll(projectId string) *Paginator[*types.Channel] {
	return &Paginator[*types.Channel]{
		c:         c.c,
		total:     -1,
		path:      "/channels",
		resultKey: "channels",
		sortBy:    "created_at",
		query:     getProjectIdParam(projectId),
	}
}

// SubscribeToken is used to subscribe a token to a channel.
func (c ClientCategoryChannels) SubscribeToken(ctx context.Context, projectId, channelId, token string) error {
	path := "/channels/" + url.PathEscape(channelId) + "/subscribers/" + url.PathEscape(token)
	return c.c.do(ctx, clientArgs{
		method:    "PUT",
		path:      path,
		query:     getProjectIdParam(projectId),
		ignore404: false,
	})
}

// SubscribeTokens is used to subscribe many tokens to a channel.
func (c ClientCategoryChannels) SubscribeTokens(ctx context.Context, projectId, channelId string, tokens []string) error {
	eg := errgroup.Group{}
	for _, v := range tokens {
		token := v
		eg.Go(func() error {
			return c.SubscribeToken(ctx, projectId, channelId, token)
		})
	}
	return eg.Wait()
}

// SetState is used to set the state of a channel.
func (c ClientCategoryChannels) SetState(ctx context.Context, projectId, id string, state map[string]any) error {
	return c.c.do(ctx, clientArgs{
		method:    "PUT",
		path:      "/channels/" + url.PathEscape(id) + "/state",
		body:      state,
		query:     getProjectIdParam(projectId),
		ignore404: false,
	})
}

// PatchState is used to patch the state of a channel.
func (c ClientCategoryChannels) PatchState(ctx context.Context, projectId, id string, state map[string]any) error {
	return c.c.do(ctx, clientArgs{
		method:    "PATCH",
		path:      "/channels/" + url.PathEscape(id) + "/state",
		body:      state,
		query:     getProjectIdParam(projectId),
		ignore404: false,
	})
}

// PublishMessage is used to publish an event to the channel.
func (c ClientCategoryChannels) PublishMessage(ctx context.Context, projectId, channelId, eventName string, data any) error {
	return c.c.do(ctx, clientArgs{
		method:    "POST",
		path:      "/channels/" + url.PathEscape(channelId) + "/messages",
		body:      map[string]any{"e": eventName, "d": data},
		query:     getProjectIdParam(projectId),
		ignore404: false,
	})
}

// Delete is used to delete a channel.
func (c ClientCategoryChannels) Delete(ctx context.Context, projectId, id string) error {
	return c.c.do(ctx, clientArgs{
		method:    "DELETE",
		path:      "/channels/" + url.PathEscape(id),
		query:     getProjectIdParam(projectId),
		ignore404: false,
	})
}

// GetStats is used to get the stats of a channel.
func (c ClientCategoryChannels) GetStats(ctx context.Context, projectId, id string) (*types.Stats, error) {
	var s types.Stats
	err := c.c.do(ctx, clientArgs{
		method:    "GET",
		path:      "/channels/" + url.PathEscape(id) + "/stats",
		resultKey: "stats",
		result:    &s,
		query:     getProjectIdParam(projectId),
		ignore404: false,
	})
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// Delete is used to delete a channel token.
func (t ClientCategoryChannelsTokens) Delete(ctx context.Context, projectId, id string) error {
	return t.c.do(ctx, clientArgs{
		method:    "DELETE",
		path:      "/channels/tokens/" + url.PathEscape(id),
		query:     getProjectIdParam(projectId),
		ignore404: false,
	})
}

// Create is used to create a new channel token. State is the map of the state of the token (this can be nil), and
// projectId is the project ID to associate the token with (this can be empty unless it is bearer or PAT auth).
func (t ClientCategoryChannelsTokens) Create(ctx context.Context, projectId string, state map[string]any) (*types.ChannelToken, error) {
	if projectId == "" && t.c.tokenType != "ptk" {
		return nil, types.InvalidToken("project ID must be specified when creating a channel token with bearer or PAT auth")
	}
	if state == nil {
		state = map[string]any{}
	}
	var ct types.ChannelToken
	err := t.c.do(ctx, clientArgs{
		method:    "POST",
		path:      "/channels/tokens",
		body:      map[string]any{"state": state},
		query:     getProjectIdParam(projectId),
		resultKey: "token",
		result:    &ct,
		ignore404: false,
	})
	if err != nil {
		return nil, err
	}
	return &ct, nil
}

// SetState is used to set the state of a channel token.
func (t ClientCategoryChannelsTokens) SetState(ctx context.Context, projectId, id string, state map[string]any) error {
	return t.c.do(ctx, clientArgs{
		method:    "PATCH",
		path:      "/channels/tokens/" + url.PathEscape(id),
		body:      map[string]any{"state": state},
		query:     getProjectIdParam(projectId),
		ignore404: false,
	})
}

// Get is used to get a token by its ID.
func (t ClientCategoryChannelsTokens) Get(ctx context.Context, projectId, id string) (*types.ChannelToken, error) {
	var ct types.ChannelToken
	err := t.c.do(ctx, clientArgs{
		method:    "GET",
		path:      "/channels/tokens/" + url.PathEscape(id),
		resultKey: "token",
		result:    &ct,
		query:     getProjectIdParam(projectId),
		ignore404: false,
	})
	if err != nil {
		return nil, err
	}
	return &ct, nil
}

// IsOnline is used to check if a token is online.
func (t ClientCategoryChannelsTokens) IsOnline(ctx context.Context, projectId, id string) (bool, error) {
	x, err := t.Get(ctx, projectId, id)
	if err != nil {
		return false, err
	}
	return x.IsOnline, nil
}

// PublishDirectMessage is used to publish an event to the channel token.
func (t ClientCategoryChannelsTokens) PublishDirectMessage(ctx context.Context, projectId, id, eventName string, data any) error {
	return t.c.do(ctx, clientArgs{
		method:    "POST",
		path:      "/channels/tokens/" + url.PathEscape(id) + "/messages",
		body:      map[string]any{"e": eventName, "d": data},
		query:     getProjectIdParam(projectId),
		ignore404: false,
	})
}

// GetAll gets all the tokens.
func (t ClientCategoryChannelsTokens) GetAll(ctx context.Context, projectId string) ([]*types.ChannelToken, error) {
	var a []*types.ChannelToken
	err := t.c.do(ctx, clientArgs{
		method:    "GET",
		path:      "/channels/tokens",
		resultKey: "tokens",
		result:    &a,
		query:     getProjectIdParam(projectId),
		ignore404: false,
	})
	if err != nil {
		return nil, err
	}
	return a, nil
}
