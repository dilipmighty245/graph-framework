package utils

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/eks"
	"sigs.k8s.io/aws-iam-authenticator/pkg/token"
)

func TestNewClientset(t *testing.T) {
	// Signing an IAM token with fixed test credentials requires no AWS requests.
	for key, value := range map[string]string{
		"AWS_ACCESS_KEY_ID": "test-access-key", "AWS_SECRET_ACCESS_KEY": "test-secret-key",
		"AWS_SESSION_TOKEN": "", "AWS_PROFILE": "", "AWS_REGION": "us-east-1",
		"AWS_SHARED_CREDENTIALS_FILE": t.TempDir() + "/credentials",
		"AWS_CONFIG_FILE":             t.TempDir() + "/config", "AWS_EC2_METADATA_DISABLED": "true",
	} {
		t.Setenv(key, value)
	}
	original, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(t.TempDir()); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(original)
	cluster := &eks.Cluster{Name: aws.String("foo"), Endpoint: aws.String("https://foo.eks.amazonaws.com"), CertificateAuthority: &eks.Certificate{Data: aws.String("LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUM1ekNDQWMrZ0F3SUJBZ0lCQURBTkJna3Foa2lHOXcwQkFRc0ZBREFWTVJNd0VRWURWUVFERXdwcmRXSmwKY201bGRHVnpNQjRYRFRJeU1Ea3hNekV3TVRjeU1Wb1hEVE15TURreE1ERXdNVGN5TVZvd0ZURVRNQkVHQTFVRQpBeE1LYTNWaVpYSnVaWFJsY3pDQ0FTSXdEUVlKS29aSWh2Y05BUUVCQlFBRGdnRVBBRENDQVFvQ2dnRUJBTnZrCjlndlZXUWxFYjhFaCtqK2RDSnpnYXE5MzFaT1RWWGZ5cldZUm93M1JhOFFCZms4OXJMRGVKQXU3aWQrendyZkcKWjFBUTVqMXFPWXUrVE5MaWducU40dURoRzREVENYZXpoN0V3TzRJTFZHTGZpYUVvdlRkZ2xjVUZYS0MvQUVvWQpBeDk4Qm43UzVmUTE1blFiUUMxWk9SczU3VVZhYWphZVBlaUJPQm05d0ljdDJrY1hWZTZMQndLVHpFTXJ4UnZaCmFJZ05hd1dqWHhVMWxINUNUazd2b0ppZW1makRtTkpZQ0dIVTVwM2NNUDF1YmVnM3RXOEFSY0hBeDNPbXJPZmEKblRvbUdlaC9MWnhlNndUUUpoRGY1WmdGMFFVaVQ1citwOG9vaGRZVTllYmZINTYxNHBuV2dJcTkrMFBiN1FYeApqYW9wUjhCcjN3YnVDcCtuNmdzQ0F3RUFBYU5DTUVBd0RnWURWUjBQQVFIL0JBUURBZ0trTUE4R0ExVWRFd0VCCi93UUZNQU1CQWY4d0hRWURWUjBPQkJZRUZKbFZMeHdRQlN6STJFSlFLK3RacitvM1poNzdNQTBHQ1NxR1NJYjMKRFFFQkN3VUFBNElCQVFDWkVNZ09KSUlFSEUxSVlkRzUreHFKK25meW1HMEpMNFNVeEpxNTZBL1hzaVcyQ2prdwpnTWNxVkVFR2tBTTVjc2FHN2h5UXFpeUk2YVl2ajNvZWVROFpDMWdnQS9LR3hXUkZYT2pPaTNOaG0ySzFBQkxqCkxMbEs4UGJidmVnVFZsQitIc3AzdUl2SFk1V2tlcU5XWGg0UFZBa1gzVzBNSmlndnlXdHE5MzEwOU1kWGkzRkQKVEdwTWFwaG9CeGU4MC9pSHg2d213RVdLMmVtakVMS1hja0J5dHYrdkpMYjkvb0VqbnpmdXZldDIrSW5Wb0tnOQp5b0NPeUFabVpLdk5uWmRtM0t1aWk2M2Z0K25TaVltZytleWZ3NkttQ0tOMlpYVDJ4WWVPMk0xd2dlL1JrZ2lUCng5dzRqbWRXWFF6TDc1QkFVQ1VmV3JnbXlQMjdmNlJ2a3grRgotLS0tLUVORCBDRVJUSUZJQ0FURS0tLS0tCg==")}}
	client, err := NewClientset(cluster)
	if err != nil || client == nil {
		t.Fatalf("client=%v error=%v", client, err)
	}
	content, err := os.ReadFile(TokenFileName)
	if err != nil || !strings.HasPrefix(string(content), "k8s-aws-v1.") {
		t.Fatalf("invalid signed token: %v", err)
	}
	info, err := os.Stat(TokenFileName)
	if err != nil || info.Mode().Perm() != 0600 {
		t.Fatal("token must be private to its owner")
	}
	cluster.CertificateAuthority.Data = aws.String("foo")
	if client, err := NewClientset(cluster); err == nil || client != nil {
		t.Fatal("invalid certificate accepted")
	}
}

func TestRefreshToken(t *testing.T) {
	for _, retry := range []bool{false, true} {
		t.Run(map[bool]string{false: "success", true: "retry"}[retry], func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			calls := 0
			refresher := TokenRefresher{TokenCh: make(chan token.Token), RefreshDuration: time.Millisecond, RetryRefreshDuration: time.Millisecond,
				FetchToken: func(*string) (token.Token, error) {
					calls++
					if retry && calls == 1 {
						return token.Token{}, errors.New("temporary credential error")
					}
					return token.Token{Token: "refreshed"}, nil
				},
			}
			done := make(chan struct{})
			go func() { RefreshToken(ctx, refresher, aws.String("foo")); close(done) }()
			select {
			case got := <-refresher.TokenCh:
				if got.Token != "refreshed" {
					t.Fatalf("unexpected token %q", got.Token)
				}
			case <-time.After(time.Second):
				t.Fatal("token was not delivered")
			}
			cancel()
			select {
			case <-done:
			case <-time.After(time.Second):
				t.Fatal("refresh did not stop")
			}
			expected := 1
			if retry {
				expected = 2
			}
			if calls != expected {
				t.Fatalf("fetches=%d, want %d", calls, expected)
			}
		})
	}
}

func TestRefreshTokenCancellation(t *testing.T) {
	for _, delay := range []time.Duration{time.Hour, 0} {
		ctx, cancel := context.WithCancel(context.Background())
		ch := make(chan token.Token)
		started := make(chan struct{})
		done := make(chan struct{})
		go func() {
			RefreshToken(ctx, TokenRefresher{TokenCh: ch, RefreshDuration: delay,
				FetchToken: func(*string) (token.Token, error) { close(started); return token.Token{}, nil },
			}, nil)
			close(done)
		}()
		if delay == 0 {
			<-started
		} // Cancel while sending to an unconsumed channel.
		cancel()
		select {
		case <-done:
		case <-time.After(time.Second):
			t.Fatal("cancellation blocked")
		}
		if _, open := <-ch; open {
			t.Fatal("channel was not closed")
		}
	}
}
