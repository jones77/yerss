package compose

import "testing"

func TestCanonicalSource(t *testing.T) {
	src := "https%3A%2F%2Fsubstack-post-media.s3.amazonaws.com%2Fpublic%2Fimages%2Fac2c7f0f-6fe9-48ad-8b7e-23d3974439ef_1600x900.jpeg"
	cases := []struct {
		name string
		in   string
		want string
	}{
		{
			"substack f_auto variant",
			"https://substackcdn.com/image/fetch/$s_!GO48!,f_auto,q_auto:good,fl_progressive:steep/" + src,
			"https://substack-post-media.s3.amazonaws.com/public/images/ac2c7f0f-6fe9-48ad-8b7e-23d3974439ef_1600x900.jpeg",
		},
		{
			"substack w_2400 variant",
			"https://substackcdn.com/image/fetch/$s_!GO48!,w_2400,c_limit,f_auto,q_auto:good,fl_progressive:steep/" + src,
			"https://substack-post-media.s3.amazonaws.com/public/images/ac2c7f0f-6fe9-48ad-8b7e-23d3974439ef_1600x900.jpeg",
		},
		{
			"substack w_1456 variant",
			"https://substackcdn.com/image/fetch/$s_!wJpA!,w_1456,c_limit,f_auto,q_auto:good,fl_progressive:steep/https%3A%2F%2Fsubstack-post-media.s3.amazonaws.com%2Fpublic%2Fimages%2Ffbf6dc44-de23-4775-a797-6b7c771523c3_1024x683.jpeg",
			"https://substack-post-media.s3.amazonaws.com/public/images/fbf6dc44-de23-4775-a797-6b7c771523c3_1024x683.jpeg",
		},
		{
			"plain URL is its own source",
			"https://cdn.example.com/photos/a.jpg",
			"https://cdn.example.com/photos/a.jpg",
		},
		{
			"http scheme encoded source",
			"https://img.example.com/fetch/http%3A%2F%2Fcdn.example.com%2Fimg%2Fb.jpg",
			"http://cdn.example.com/img/b.jpg",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := CanonicalSource(c.in); got != c.want {
				t.Errorf("CanonicalSource(%q) = %q, want %q", c.in, got, c.want)
			}
		})
	}
}
