package determinator

// compile time check that all retrievers implement the retriever interface
var _ Retriever = &FileRetriever{}
var _ Retriever = &CachedRetriever{}
var _ Retriever = &HTTPRetriever{}
var _ Retriever = &MockRetriever{}
var _ Retriever = &TrackingRetriever{}
