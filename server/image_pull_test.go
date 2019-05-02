package server_test

import (
	"context"

	"github.com/containers/image/types"
	"github.com/cri-o/cri-o/pkg/storage"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	digest "github.com/opencontainers/go-digest"
	pb "k8s.io/kubernetes/pkg/kubelet/apis/cri/runtime/v1alpha2"
)

// The actual test suite
var _ = t.Describe("ImagePull", func() {
	// Prepare the sut
	BeforeEach(func() {
		beforeEach()
		setupSUT()
	})
	AfterEach(afterEach)

	t.Describe("ImagePull", func() {
		It("should succeed with pull", func() {
			// Given
			gomock.InOrder(
				imageServerMock.EXPECT().ResolveNames(gomock.Any()).
					Return([]string{"image"}, nil),
				imageServerMock.EXPECT().PrepareImage(gomock.Any(),
					gomock.Any()).Return(imageMock, nil),
				imageServerMock.EXPECT().ImageStatus(
					gomock.Any(), gomock.Any()).
					Return(&storage.ImageResult{ID: "image"}, nil),
				imageMock.EXPECT().ConfigInfo().
					Return(types.BlobInfo{Digest: digest.Digest("")}),
				imageServerMock.EXPECT().PullImage(
					gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, nil),
				imageServerMock.EXPECT().ImageStatus(
					gomock.Any(), gomock.Any()).
					Return(&storage.ImageResult{ID: "image",
						RepoDigests: []string{"digest"}}, nil),
			)

			// When
			response, err := sut.PullImage(context.Background(),
				&pb.PullImageRequest{Image: &pb.ImageSpec{
					Image: "id"}})

			// Then
			Expect(err).To(BeNil())
			Expect(response).NotTo(BeNil())
		})

		It("should succeed when already pulled", func() {
			// Given
			gomock.InOrder(
				imageServerMock.EXPECT().ResolveNames(gomock.Any()).
					Return([]string{"image"}, nil),
				imageServerMock.EXPECT().PrepareImage(gomock.Any(),
					gomock.Any()).Return(imageMock, nil),
				imageServerMock.EXPECT().ImageStatus(
					gomock.Any(), gomock.Any()).
					Return(&storage.ImageResult{ID: "image",
						ConfigDigest: digest.Digest("digest")}, nil),
				imageMock.EXPECT().ConfigInfo().
					Return(types.BlobInfo{Digest: digest.Digest("digest")}),
				imageServerMock.EXPECT().ImageStatus(
					gomock.Any(), gomock.Any()).
					Return(&storage.ImageResult{ID: "image",
						RepoDigests: []string{"digest"}}, nil),
			)

			// When
			response, err := sut.PullImage(context.Background(),
				&pb.PullImageRequest{
					Image: &pb.ImageSpec{Image: "id"},
					Auth: &pb.AuthConfig{
						Username: "username",
						Password: "password",
						Auth:     "auth",
					},
				})

			// Then
			Expect(err).To(BeNil())
			Expect(response).NotTo(BeNil())
		})

		It("should fail when second image status retrieval errors", func() {
			// Given
			gomock.InOrder(
				imageServerMock.EXPECT().ResolveNames(gomock.Any()).
					Return([]string{"image"}, nil),
				imageServerMock.EXPECT().PrepareImage(gomock.Any(),
					gomock.Any()).Return(imageMock, nil),
				imageServerMock.EXPECT().ImageStatus(
					gomock.Any(), gomock.Any()).
					Return(&storage.ImageResult{ID: "image"}, nil),
				imageMock.EXPECT().ConfigInfo().
					Return(types.BlobInfo{Digest: digest.Digest("")}),
				imageServerMock.EXPECT().PullImage(
					gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, nil),
				imageServerMock.EXPECT().ImageStatus(
					gomock.Any(), gomock.Any()).
					Return(nil, t.TestError),
			)

			// When
			response, err := sut.PullImage(context.Background(),
				&pb.PullImageRequest{
					Image: &pb.ImageSpec{Image: "id"},
					Auth: &pb.AuthConfig{
						Username: "username",
						Password: "password",
						Auth:     "YWJjOmFiYw==",
					},
				})

			// Then
			Expect(err).NotTo(BeNil())
			Expect(response).To(BeNil())
		})

		It("should fail credential decode erros", func() {
			// Given
			gomock.InOrder(
				imageServerMock.EXPECT().ResolveNames(gomock.Any()).
					Return([]string{"image"}, nil),
			)

			// When
			response, err := sut.PullImage(context.Background(),
				&pb.PullImageRequest{
					Image: &pb.ImageSpec{Image: "id"},
					Auth: &pb.AuthConfig{
						Auth: "❤️",
					},
				})

			// Then
			Expect(err).NotTo(BeNil())
			Expect(response).To(BeNil())
		})

		It("should fail when image pull erros", func() {
			// Given
			gomock.InOrder(
				imageServerMock.EXPECT().ResolveNames(gomock.Any()).
					Return([]string{"image"}, nil),
				imageServerMock.EXPECT().PrepareImage(gomock.Any(),
					gomock.Any()).Return(imageMock, nil),
				imageServerMock.EXPECT().ImageStatus(
					gomock.Any(), gomock.Any()).
					Return(&storage.ImageResult{ID: "image"}, nil),
				imageMock.EXPECT().ConfigInfo().
					Return(types.BlobInfo{Digest: digest.Digest("")}),
				imageServerMock.EXPECT().PullImage(
					gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, t.TestError),
			)

			// When
			response, err := sut.PullImage(context.Background(),
				&pb.PullImageRequest{Image: &pb.ImageSpec{
					Image: "id"}})

			// Then
			Expect(err).NotTo(BeNil())
			Expect(response).To(BeNil())
		})

		It("should fail when prepare image errors", func() {
			// Given
			gomock.InOrder(
				imageServerMock.EXPECT().ResolveNames(gomock.Any()).
					Return([]string{"image"}, nil),
				imageServerMock.EXPECT().PrepareImage(gomock.Any(),
					gomock.Any()).Return(nil, t.TestError),
			)

			// When
			response, err := sut.PullImage(context.Background(),
				&pb.PullImageRequest{Image: &pb.ImageSpec{
					Image: "id"}})

			// Then
			Expect(err).NotTo(BeNil())
			Expect(response).To(BeNil())
		})

		It("should fail when resolve names errors", func() {
			// Given
			gomock.InOrder(
				imageServerMock.EXPECT().ResolveNames(gomock.Any()).
					Return(nil, t.TestError),
			)
			// When
			response, err := sut.PullImage(context.Background(),
				&pb.PullImageRequest{})

			// Then
			Expect(err).NotTo(BeNil())
			Expect(response).To(BeNil())
		})

	})
})
